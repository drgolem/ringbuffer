// Package ringbuffer provides a lock-free SPSC (Single Producer Single Consumer) ring buffer.
//
// This implementation is optimized for scenarios where one goroutine writes (producer)
// and another goroutine reads (consumer), such as audio streaming, network I/O buffering,
// or any producer-consumer pattern requiring high-throughput lock-free communication.
//
// # Thread Safety
//
// This ring buffer is ONLY safe for single producer + single consumer scenarios.
// It uses atomic operations to achieve thread-safety without locks, providing
// excellent performance for SPSC use cases.
//
// IMPORTANT: Multiple producers or multiple consumers will cause data races.
//
// # Features
//
//   - Lock-free SPSC implementation using atomic operations
//   - Zero-copy access via ReadSlices() and PeekContiguous()
//   - Standard io.Reader and io.Writer interface implementations
//   - Power-of-2 sizing for efficient modulo operations
//   - Production-tested in high-performance audio streaming (>200k frames/sec)
//
// # Basic Usage
//
//	rb := ringbuffer.New(1024)
//
//	// Producer goroutine
//	go func() {
//	    data := []byte("hello")
//	    rb.Write(data)
//	}()
//
//	// Consumer goroutine
//	buffer := make([]byte, 100)
//	n, err := rb.Read(buffer)
//
// # Zero-Copy Usage
//
//	// Get direct access to buffer contents
//	first, second, total := rb.ReadSlices()
//	// Process first and second slices...
//	rb.Consume(total)  // Advance read position
package ringbuffer

import (
	"errors"
	"sync/atomic"
)

// Common ringbuffer errors used for error handling and comparison using errors.Is().
var (
	// ErrInsufficientSpace indicates the ringbuffer doesn't have enough space for the write operation
	ErrInsufficientSpace = errors.New("insufficient space in ringbuffer")

	// ErrInsufficientData indicates the ringbuffer doesn't have enough data for the read operation
	ErrInsufficientData = errors.New("insufficient data in ringbuffer")
)

// RingBuffer is a lock-free single-producer single-consumer ring buffer
// optimized for high-throughput byte data streaming.
//
// RingBuffer implements io.Reader and io.Writer interfaces, making it compatible
// with the standard Go I/O ecosystem. However, note the thread safety requirements:
//   - Write() must only be called by the producer thread (implements io.Writer)
//   - Read() must only be called by the consumer thread (implements io.Reader)
//
// The implementation uses atomic operations for lock-free synchronization between
// producer and consumer, achieving high performance without mutex overhead.
type RingBuffer struct {
	buffer   []byte
	size     uint64 // must be power of 2
	mask     uint64 // size - 1, for efficient modulo
	writePos atomic.Uint64
	readPos  atomic.Uint64
}

// New creates a new ring buffer with the given size.
// Size will be rounded up to the next power of 2 for efficiency.
//
// Power-of-2 sizing enables fast modulo operations using bitwise AND,
// which is critical for performance in high-throughput scenarios.
//
// Example:
//
//	rb := ringbuffer.New(1000)  // Creates buffer with size 1024 (next power of 2)
func New(size uint64) *RingBuffer {
	// Round up to next power of 2
	size = nextPowerOf2(size)

	return &RingBuffer{
		buffer: make([]byte, size),
		size:   size,
		mask:   size - 1,
	}
}

// Write writes data to the ring buffer, implementing io.Writer.
// It writes all of len(data) bytes or returns an error.
//
// Unlike some io.Writer implementations, this method does not perform partial writes.
// It will either write all data successfully or return ErrInsufficientSpace without
// writing any data.
//
// This method must only be called by the producer thread.
//
// Returns:
//   - Number of bytes written (always len(data) or 0)
//   - ErrInsufficientSpace if buffer doesn't have room for all data
func (rb *RingBuffer) Write(data []byte) (int, error) {
	dataLen := uint64(len(data))
	if dataLen == 0 {
		return 0, nil
	}

	available := rb.AvailableWrite()
	if dataLen > available {
		return 0, ErrInsufficientSpace
	}

	writePos := rb.writePos.Load()

	// Calculate the actual position in the buffer
	start := writePos & rb.mask
	end := (writePos + dataLen) & rb.mask

	if end > start {
		// Single contiguous write
		copy(rb.buffer[start:end], data)
	} else {
		// Write wraps around the buffer
		firstChunk := rb.size - start
		copy(rb.buffer[start:], data[:firstChunk])
		copy(rb.buffer[:end], data[firstChunk:])
	}

	// Atomic update of write position
	rb.writePos.Store(writePos + dataLen)

	return int(dataLen), nil
}

// Read reads up to len(data) bytes from the ring buffer into data, implementing io.Reader.
//
// Read will read as many bytes as are available, up to len(data). If fewer bytes are
// available than requested, it reads what's available and returns the count without error.
// If the buffer is empty, it returns (0, ErrInsufficientData).
//
// This follows io.Reader semantics where ErrInsufficientData is analogous to io.EOF,
// indicating no data is currently available for reading.
//
// This method must only be called by the consumer thread.
//
// Returns:
//   - Number of bytes actually read
//   - ErrInsufficientData if buffer is empty, nil otherwise
func (rb *RingBuffer) Read(data []byte) (int, error) {
	dataLen := uint64(len(data))
	if dataLen == 0 {
		return 0, nil
	}

	available := rb.AvailableRead()
	if available == 0 {
		return 0, ErrInsufficientData
	}

	// Read only what's available
	toRead := min(dataLen, available)

	readPos := rb.readPos.Load()

	// Calculate the actual position in the buffer
	start := readPos & rb.mask
	end := (readPos + toRead) & rb.mask

	if end > start {
		// Single contiguous read
		copy(data[:toRead], rb.buffer[start:end])
	} else {
		// Read wraps around the buffer
		firstChunk := rb.size - start
		copy(data[:firstChunk], rb.buffer[start:])
		copy(data[firstChunk:toRead], rb.buffer[:end])
	}

	// Atomic update of read position
	rb.readPos.Store(readPos + toRead)

	return int(toRead), nil
}

// AvailableWrite returns the number of bytes available for writing.
// Safe to call from producer thread.
func (rb *RingBuffer) AvailableWrite() uint64 {
	writePos := rb.writePos.Load()
	readPos := rb.readPos.Load()
	return rb.size - (writePos - readPos)
}

// AvailableRead returns the number of bytes available for reading.
// Safe to call from consumer thread.
func (rb *RingBuffer) AvailableRead() uint64 {
	writePos := rb.writePos.Load()
	readPos := rb.readPos.Load()
	return writePos - readPos
}

// Size returns the total size of the ring buffer.
// This is the capacity, not the number of bytes currently stored.
func (rb *RingBuffer) Size() uint64 {
	return rb.size
}

// ReadSlices returns one or two slices that provide zero-copy access to the available data.
// The data may be split into two slices if it wraps around the ring buffer.
// After processing the data, call Consume() to advance the read position.
// This should only be called by the consumer thread.
//
// This is a zero-copy operation - no data is copied, only pointers are returned.
// This is ideal for high-performance scenarios where you can process data in-place.
//
// Returns:
//   - first: The first (or only) slice of available data
//   - second: The second slice if data wraps around, nil otherwise
//   - total: Total number of bytes available across both slices
//
// Example:
//
//	first, second, total := rb.ReadSlices()
//	if total > 0 {
//	    process(first)
//	    if second != nil {
//	        process(second)
//	    }
//	    rb.Consume(total)
//	}
func (rb *RingBuffer) ReadSlices() (first, second []byte, total uint64) {
	available := rb.AvailableRead()
	if available == 0 {
		return nil, nil, 0
	}

	readPos := rb.readPos.Load()
	start := readPos & rb.mask
	end := (readPos + available) & rb.mask

	if end > start {
		// Data is contiguous
		return rb.buffer[start:end], nil, available
	}

	// Data wraps around
	firstChunk := rb.buffer[start:]
	secondChunk := rb.buffer[:end]
	return firstChunk, secondChunk, available
}

// PeekContiguous returns a slice providing zero-copy access to the contiguous
// portion of available data. This may be less than the total available data
// if the data wraps around the buffer.
// After processing, call Consume() to advance the read position.
// This should only be called by the consumer thread.
//
// This is useful when you need a single contiguous slice and don't want to
// handle the wrap-around case.
//
// Example:
//
//	data := rb.PeekContiguous()
//	if len(data) > 0 {
//	    n := process(data)  // Process what you can
//	    rb.Consume(uint64(n))
//	}
func (rb *RingBuffer) PeekContiguous() []byte {
	available := rb.AvailableRead()
	if available == 0 {
		return nil
	}

	readPos := rb.readPos.Load()
	start := readPos & rb.mask
	end := (readPos + available) & rb.mask

	if end > start {
		// All data is contiguous
		return rb.buffer[start:end]
	}

	// Data wraps around, return only the first contiguous chunk
	return rb.buffer[start:]
}

// Consume advances the read position by n bytes without copying data.
// This is used in conjunction with ReadSlices() or PeekContiguous() for zero-copy reads.
// Returns an error if trying to consume more bytes than are available.
// This should only be called by the consumer thread.
//
// Example:
//
//	data := rb.PeekContiguous()
//	processed := doSomething(data)
//	rb.Consume(uint64(processed))
func (rb *RingBuffer) Consume(n uint64) error {
	if n == 0 {
		return nil
	}

	available := rb.AvailableRead()
	if n > available {
		return ErrInsufficientData
	}

	readPos := rb.readPos.Load()
	rb.readPos.Store(readPos + n)
	return nil
}

// Reset clears the ring buffer by resetting read and write positions.
// This should only be called when both producer and consumer are idle,
// as it is not safe to call concurrently with Read() or Write().
func (rb *RingBuffer) Reset() {
	rb.readPos.Store(0)
	rb.writePos.Store(0)
}

// nextPowerOf2 rounds up to the next power of 2.
// This enables efficient modulo operations using bitwise AND.
func nextPowerOf2(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return n
}

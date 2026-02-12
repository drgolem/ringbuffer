package ringbuffer

import (
	"fmt"
)

// RingBuffer is a circular buffer implementation for byte data.
// It provides FIFO (First-In-First-Out) semantics with fixed capacity.
// This implementation is not thread-safe.
type RingBuffer struct {
	buf      []byte
	readPos  int
	writePos int
	size     int
}

// NewRingBuffer creates a new ring buffer with the specified capacity.
// Returns an error if capacity is not positive.
func NewRingBuffer(capacity int) (*RingBuffer, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0, got %d", capacity)
	}

	rb := &RingBuffer{
		buf: make([]byte, capacity),
	}

	return rb, nil
}

// Capacity returns the total capacity of the ring buffer.
func (rb *RingBuffer) Capacity() int {
	return len(rb.buf)
}

// Size returns the current number of bytes stored in the buffer.
func (rb *RingBuffer) Size() int {
	return rb.size
}

// AvailableWriteSize returns the number of bytes available for writing.
func (rb *RingBuffer) AvailableWriteSize() int {
	return len(rb.buf) - rb.size
}

// IsEmpty returns true if the buffer contains no data.
func (rb *RingBuffer) IsEmpty() bool {
	return rb.size == 0
}

// IsFull returns true if the buffer is at full capacity.
func (rb *RingBuffer) IsFull() bool {
	return rb.size == len(rb.buf)
}

// Reset clears the buffer, removing all data and resetting read/write positions.
func (rb *RingBuffer) Reset() {
	rb.size = 0
	rb.readPos = 0
	rb.writePos = 0
}

// Read reads n bytes from the ring buffer into dst.
// Returns the number of bytes read and an error if the operation fails.
// Returns an error if n is negative, dst is too small, or insufficient data is available.
func (rb *RingBuffer) Read(n int, dst []byte) (int, error) {
	if n < 0 {
		return 0, fmt.Errorf("n must be non-negative, got %d", n)
	}
	if n > len(dst) {
		return 0, fmt.Errorf("destination buffer too small: need %d bytes, have %d", n, len(dst))
	}
	if rb.size < n {
		return 0, fmt.Errorf("insufficient data in buffer: need %d bytes, have %d", n, rb.size)
	}

	capacity := len(rb.buf)
	if rb.readPos+n <= capacity {
		copy(dst, rb.buf[rb.readPos:rb.readPos+n])
		rb.readPos = (rb.readPos + n) % capacity
		rb.size -= n
		return n, nil
	}

	copy(dst, rb.buf[rb.readPos:capacity])
	copy(dst[capacity-rb.readPos:], rb.buf[0:n-(capacity-rb.readPos)])

	rb.readPos = (n - (capacity - rb.readPos)) % capacity
	rb.size -= n

	return n, nil
}

// Write writes data to the ring buffer.
// Returns the number of bytes written and an error if the operation fails.
// Returns an error if there is insufficient space in the buffer.
func (rb *RingBuffer) Write(data []byte) (int, error) {
	szData := len(data)
	capacity := len(rb.buf)
	if szData > capacity-rb.size {
		return 0, fmt.Errorf("data length exceeds available write space: need %d bytes, have %d", szData, capacity-rb.size)
	}

	szToEnd := rb.writePos + szData
	if szToEnd <= capacity {
		copy(rb.buf[rb.writePos:], data)
		rb.writePos = (rb.writePos + szData) % capacity
	} else {
		sz1 := capacity - rb.writePos
		copy(rb.buf[rb.writePos:], data[:sz1])
		copy(rb.buf[0:], data[sz1:])
		rb.writePos = (szData - sz1) % capacity
	}
	rb.size += szData

	return szData, nil
}

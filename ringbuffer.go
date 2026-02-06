package ringbuffer

import (
	"fmt"
)

type RingBuffer struct {
	buf      []byte
	readPos  int
	writePos int
	size     int
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		panic("capacity must be positive")
	}

	rb := RingBuffer{
		buf: make([]byte, capacity),
	}

	return &rb
}

func (rb *RingBuffer) Capacity() int {
	return len(rb.buf)
}

func (rb *RingBuffer) Size() int {
	return rb.size
}

func (rb *RingBuffer) AvailableWriteSize() int {
	return rb.Capacity() - rb.Size()
}

func (rb *RingBuffer) IsEmpty() bool {
	return rb.Size() == 0
}

func (rb *RingBuffer) IsFull() bool {
	return rb.Size() == rb.Capacity()
}

func (rb *RingBuffer) Reset() {
	rb.size = 0
	rb.readPos = 0
	rb.writePos = 0
}

func (rb *RingBuffer) Read(n int, dst []byte) (int, error) {
	if rb.size < n {
		return 0, fmt.Errorf("invalid destination buffer size. sz: %d, n: %d", rb.size, n)
	}

	if rb.readPos+n <= rb.Capacity() {
		copy(dst, rb.buf[rb.readPos:rb.readPos+n])
		rb.readPos = (rb.readPos + n) % rb.Capacity()
		rb.size -= n
		return n, nil
	}

	copy(dst, rb.buf[rb.readPos:rb.Capacity()])
	copy(dst[rb.Capacity()-rb.readPos:], rb.buf[0:n-(rb.Capacity()-rb.readPos)])

	rb.readPos = (n - (rb.Capacity() - rb.readPos)) % rb.Capacity()
	rb.size -= n

	return n, nil
}

func (rb *RingBuffer) Write(data []byte) (int, error) {
	szData := len(data)
	if szData > rb.Capacity()-rb.Size() {
		return 0, fmt.Errorf("data len exceeds available write size: %d > %d", szData, rb.Capacity()-rb.Size())
	}

	szToEnd := rb.writePos + szData
	if szToEnd <= rb.Capacity() {
		copy(rb.buf[rb.writePos:], data)
		rb.writePos = (rb.writePos + szData) % rb.Capacity()
	} else {
		sz1 := rb.Capacity() - rb.writePos
		copy(rb.buf[rb.writePos:], data[:sz1])
		copy(rb.buf[0:], data[sz1:])
		rb.writePos = (szData - sz1) % rb.Capacity()
	}
	rb.size += szData

	return szData, nil
}

package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRingBuffer(t *testing.T) {
	capacity := 5

	rb, err := NewRingBuffer(capacity)
	assert.NoError(t, err)
	assert.NotNil(t, rb)
	assert.Equal(t, capacity, rb.Capacity())
	assert.Equal(t, 0, rb.Size())
}

func TestNewRingBufferInvalidCapacity(t *testing.T) {
	// Test zero capacity
	rb, err := NewRingBuffer(0)
	assert.Error(t, err)
	assert.Nil(t, rb)

	// Test negative capacity
	rb, err = NewRingBuffer(-5)
	assert.Error(t, err)
	assert.Nil(t, rb)
}

func TestWriteBasic(t *testing.T) {
	capacity := 5

	rb, err := NewRingBuffer(capacity)
	assert.NoError(t, err)

	data := []byte{1, 2, 3}

	nw, err := rb.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), nw)
	assert.Equal(t, nw, rb.Size())

	buf_expect := []byte{1, 2, 3, 0, 0}
	assert.Equal(t, buf_expect, rb.buf)
}

func TestWriteExceedsCapacity(t *testing.T) {
	capacity := 5

	rb, err := NewRingBuffer(capacity)
	assert.NoError(t, err)

	data1 := []byte{1, 2, 3}

	nw, err := rb.Write(data1)
	assert.NoError(t, err)
	assert.Equal(t, len(data1), nw)
	assert.Equal(t, 3, rb.Size())

	data2 := []byte{4, 5}

	nw, err = rb.Write(data2)
	assert.NoError(t, err)
	assert.Equal(t, len(data2), nw)
	assert.Equal(t, 5, rb.Size())

	buf_expect := []byte{1, 2, 3, 4, 5}
	assert.Equal(t, buf_expect, rb.buf)

	data3 := []byte{6, 7}

	nw, err = rb.Write(data3)
	assert.Error(t, err)
}

func TestReadBasic(t *testing.T) {
	capacity := 5

	rb, err := NewRingBuffer(capacity)
	assert.NoError(t, err)

	data1 := []byte{1, 2, 3}

	nw, err := rb.Write(data1)
	assert.NoError(t, err)
	assert.Equal(t, len(data1), nw)
	assert.Equal(t, 3, rb.Size())

	out1 := make([]byte, 3)
	nr, err := rb.Read(2, out1)
	assert.NoError(t, err)
	assert.Equal(t, 2, nr)
	assert.Equal(t, 1, rb.Size())

	assert.Equal(t, []byte{1, 2}, out1[:2])
	assert.Equal(t, 3, rb.writePos)

	nr, err = rb.Read(2, out1)
	assert.Error(t, err)

	nr, err = rb.Read(1, out1)
	assert.NoError(t, err)
	assert.Equal(t, 1, nr)
	assert.Equal(t, 0, rb.Size())
}

func TestReadWriteFullCapacity(t *testing.T) {
	capacity := 5

	rb, err := NewRingBuffer(capacity)
	assert.NoError(t, err)

	data1 := []byte{1, 2, 3, 4, 5}

	nw, err := rb.Write(data1)
	assert.NoError(t, err)
	assert.Equal(t, len(data1), nw)
	assert.Equal(t, 5, rb.Size())
	assert.Equal(t, 0, rb.writePos)

	out1 := make([]byte, 1)
	nr, err := rb.Read(1, out1)
	assert.NoError(t, err)
	assert.Equal(t, 1, nr)
	assert.Equal(t, 4, rb.Size())
	assert.Equal(t, []byte{1}, out1[:1])

	out2 := make([]byte, 4)
	nr, err = rb.Read(4, out2)
	assert.NoError(t, err)
	assert.Equal(t, 4, nr)
	assert.Equal(t, 0, rb.Size())
	assert.Equal(t, []byte{2, 3, 4, 5}, out2[:4])
	assert.Equal(t, 0, rb.readPos)
}

func TestIsEmpty(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	// Should be empty initially
	assert.True(t, rb.IsEmpty())

	// Should not be empty after writing
	_, err = rb.Write([]byte{1, 2})
	assert.NoError(t, err)
	assert.False(t, rb.IsEmpty())

	// Should be empty after reading all data
	out := make([]byte, 2)
	_, err = rb.Read(2, out)
	assert.NoError(t, err)
	assert.True(t, rb.IsEmpty())
}

func TestIsFull(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	// Should not be full initially
	assert.False(t, rb.IsFull())

	// Should be full after writing to capacity
	_, err = rb.Write([]byte{1, 2, 3, 4, 5})
	assert.NoError(t, err)
	assert.True(t, rb.IsFull())

	// Should not be full after reading some data
	out := make([]byte, 2)
	_, err = rb.Read(2, out)
	assert.NoError(t, err)
	assert.False(t, rb.IsFull())
}

func TestAvailableWriteSize(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	// Should have full capacity available initially
	assert.Equal(t, 5, rb.AvailableWriteSize())

	// Should have reduced space after writing
	_, err = rb.Write([]byte{1, 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, 2, rb.AvailableWriteSize())

	// Should have more space after reading
	out := make([]byte, 2)
	_, err = rb.Read(2, out)
	assert.NoError(t, err)
	assert.Equal(t, 4, rb.AvailableWriteSize())
}

func TestReset(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	// Write some data
	_, err = rb.Write([]byte{1, 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, 3, rb.Size())

	// Reset should clear everything
	rb.Reset()
	assert.Equal(t, 0, rb.Size())
	assert.Equal(t, 0, rb.readPos)
	assert.Equal(t, 0, rb.writePos)
	assert.True(t, rb.IsEmpty())
}

func TestReadNegativeN(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	_, err = rb.Write([]byte{1, 2, 3})
	assert.NoError(t, err)

	out := make([]byte, 3)
	_, err = rb.Read(-1, out)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative")
}

func TestReadBufferTooSmall(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	_, err = rb.Write([]byte{1, 2, 3})
	assert.NoError(t, err)

	out := make([]byte, 2)
	_, err = rb.Read(3, out)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too small")
}

func TestReadInsufficientData(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	_, err = rb.Write([]byte{1, 2})
	assert.NoError(t, err)

	out := make([]byte, 5)
	_, err = rb.Read(3, out)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestWriteWraparound(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	// Write 3 bytes
	_, err = rb.Write([]byte{1, 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, 3, rb.writePos)

	// Read 2 bytes
	out := make([]byte, 2)
	_, err = rb.Read(2, out)
	assert.NoError(t, err)
	assert.Equal(t, 2, rb.readPos)

	// Write 4 bytes (will wrap around)
	_, err = rb.Write([]byte{4, 5, 6, 7})
	assert.NoError(t, err)
	assert.Equal(t, 2, rb.writePos) // Should wrap to position 2
	assert.Equal(t, 5, rb.Size())

	// Verify buffer contents
	expected := []byte{6, 7, 3, 4, 5}
	assert.Equal(t, expected, rb.buf)
}

func TestReadWraparound(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	// Write 3 bytes
	_, err = rb.Write([]byte{1, 2, 3})
	assert.NoError(t, err)

	// Read 2 bytes
	out := make([]byte, 2)
	_, err = rb.Read(2, out)
	assert.NoError(t, err)
	assert.Equal(t, []byte{1, 2}, out)

	// Write 4 more bytes (wraps around)
	_, err = rb.Write([]byte{4, 5, 6, 7})
	assert.NoError(t, err)

	// Read all 5 bytes (should wrap around)
	out = make([]byte, 5)
	_, err = rb.Read(5, out)
	assert.NoError(t, err)
	assert.Equal(t, []byte{3, 4, 5, 6, 7}, out)
	assert.Equal(t, 0, rb.Size())
}

func TestMultipleReadWriteCycles(t *testing.T) {
	rb, err := NewRingBuffer(8)
	assert.NoError(t, err)

	for i := 0; i < 100; i++ {
		// Write 3 bytes
		data := []byte{byte(i), byte(i + 1), byte(i + 2)}
		_, err = rb.Write(data)
		assert.NoError(t, err)

		// Read 3 bytes
		out := make([]byte, 3)
		_, err = rb.Read(3, out)
		assert.NoError(t, err)
		assert.Equal(t, data, out)
	}

	assert.Equal(t, 0, rb.Size())
}

func TestWriteEmptySlice(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	nw, err := rb.Write([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, 0, nw)
	assert.Equal(t, 0, rb.Size())
}

func TestReadZeroBytes(t *testing.T) {
	rb, err := NewRingBuffer(5)
	assert.NoError(t, err)

	_, err = rb.Write([]byte{1, 2, 3})
	assert.NoError(t, err)

	out := make([]byte, 5)
	nr, err := rb.Read(0, out)
	assert.NoError(t, err)
	assert.Equal(t, 0, nr)
	assert.Equal(t, 3, rb.Size()) // Size should not change
}

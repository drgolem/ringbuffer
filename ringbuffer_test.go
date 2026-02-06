package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRingBuffer(t *testing.T) {

	capacity := 5

	rb := NewRingBuffer(capacity)

	assert.Equal(t, capacity, rb.Capacity())
	assert.Equal(t, 0, rb.Size())
}

func TestWriteBasic(t *testing.T) {

	capacity := 5

	rb := NewRingBuffer(capacity)

	data := []byte{1, 2, 3}

	nw, err := rb.Write(data)
	assert.Nil(t, err)
	assert.Equal(t, nw, len(data))
	assert.Equal(t, nw, rb.Size())

	buf_expect := []byte{1, 2, 3, 0, 0}
	assert.Equal(t, buf_expect, rb.buf)
}

func TestWriteExceedsCapacity(t *testing.T) {
	capacity := 5

	rb := NewRingBuffer(capacity)

	data1 := []byte{1, 2, 3}

	nw, err := rb.Write(data1)
	assert.Nil(t, err)
	assert.Equal(t, nw, len(data1))
	assert.Equal(t, 3, rb.Size())

	data2 := []byte{4, 5}

	nw, err = rb.Write(data2)
	assert.Nil(t, err)
	assert.Equal(t, nw, len(data2))
	assert.Equal(t, 5, rb.Size())

	buf_expect := []byte{1, 2, 3, 4, 5}
	assert.Equal(t, buf_expect, rb.buf)

	data3 := []byte{6, 7}

	nw, err = rb.Write(data3)
	assert.NotNil(t, err)
}

func TestReadBasic(t *testing.T) {

	capacity := 5

	rb := NewRingBuffer(capacity)

	data1 := []byte{1, 2, 3}

	nw, err := rb.Write(data1)
	assert.Nil(t, err)
	assert.Equal(t, nw, len(data1))
	assert.Equal(t, 3, rb.Size())

	out1 := make([]byte, 3)
	nr, err := rb.Read(2, out1)
	assert.Nil(t, err)
	assert.Equal(t, 2, nr)
	assert.Equal(t, 1, rb.Size())

	assert.Equal(t, []byte{1, 2}, out1[:2])
	assert.Equal(t, 3, rb.writePos)

	nr, err = rb.Read(2, out1)
	assert.NotNil(t, err)

	nr, err = rb.Read(1, out1)
	assert.Nil(t, err)
	assert.Equal(t, 1, nr)
	assert.Equal(t, 0, rb.Size())
}

func TestReadWriteFullCapacity(t *testing.T) {
	capacity := 5

	rb := NewRingBuffer(capacity)

	data1 := []byte{1, 2, 3, 4, 5}

	nw, err := rb.Write(data1)
	assert.Nil(t, err)
	assert.Equal(t, nw, len(data1))
	assert.Equal(t, 5, rb.Size())
	assert.Equal(t, 0, rb.writePos)

	out1 := make([]byte, 1)
	nr, err := rb.Read(1, out1)
	assert.Nil(t, err)
	assert.Equal(t, 1, nr)
	assert.Equal(t, 4, rb.Size())
	assert.Equal(t, []byte{1}, out1[:1])

	out2 := make([]byte, 4)
	nr, err = rb.Read(4, out2)
	assert.Nil(t, err)
	assert.Equal(t, 4, nr)
	assert.Equal(t, 0, rb.Size())
	assert.Equal(t, []byte{2, 3, 4, 5}, out2[:4])
	assert.Equal(t, 0, rb.readPos)
}

package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_1(t *testing.T) {

	capacity := 5

	rb := NewRingBuffer(capacity)

	assert.Equal(t, capacity, rb.Capacity())
	assert.Equal(t, 0, rb.Size())
}

func Test_2(t *testing.T) {

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

func Test_3(t *testing.T) {

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

func Test_4(t *testing.T) {

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

func Test_5(t *testing.T) {

	capacity := 5

	rb := NewRingBuffer(capacity)

	data1 := []byte{1, 2, 3}

	nw, err := rb.Write(data1)
	assert.Nil(t, err)
	assert.Equal(t, nw, len(data1))
	assert.Equal(t, 3, rb.Size())
	assert.Equal(t, 3, rb.writePos)

	out1 := make([]byte, 3)
	nr, err := rb.Read(2, out1)
	assert.Nil(t, err)
	assert.Equal(t, 2, nr)
	assert.Equal(t, 1, rb.Size())
	assert.Equal(t, []byte{1, 2}, out1[:2])

	data2 := []byte{4, 5, 6}
	assert.Equal(t, 3, rb.writePos)
	nw, err = rb.Write(data2)
	assert.Nil(t, err)
	assert.Equal(t, nw, len(data2))
	assert.Equal(t, 4, rb.Size())
	assert.Equal(t, 1, rb.writePos)
	buf_expect := []byte{6, 2, 3, 4, 5}
	assert.Equal(t, buf_expect, rb.buf)

	assert.Equal(t, 2, rb.readPos)
	out2 := make([]byte, 4)
	nr, err = rb.Read(4, out2)
	assert.Nil(t, err)
	assert.Equal(t, 4, nr)
	assert.Equal(t, 0, rb.Size())
	assert.Equal(t, []byte{3, 4, 5, 6}, out2[:4])
}

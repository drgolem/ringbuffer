# ringbuffer

A lock-free SPSC (Single Producer Single Consumer) ring buffer implementation in Go.

## Features

- Lock-free implementation using atomic operations
- Supports single producer and single consumer (SPSC)
- Implements `io.Reader` and `io.Writer` interfaces
- Zero-copy read operations available
- Buffer size rounded to next power of 2

## Installation

```bash
go get github.com/drgolem/ringbuffer
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/drgolem/ringbuffer"
)

func main() {
    rb := ringbuffer.New(1024)

    // Write data
    data := []byte("Hello, World!")
    n, err := rb.Write(data)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Wrote %d bytes\n", n)

    // Read data
    buf := make([]byte, 100)
    n, err = rb.Read(buf)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Read: %s\n", buf[:n])
}
```

## API

### Constructor

- `New(size uint64) *RingBuffer` - Creates a new ring buffer

### Reading

- `Read(data []byte) (int, error)` - Reads data from buffer
- `ReadSlices() ([]byte, []byte, uint64)` - Returns direct buffer slices (zero-copy)
- `PeekContiguous() []byte` - Returns contiguous readable data (zero-copy)
- `Consume(n uint64) error` - Advances read position

### Writing

- `Write(data []byte) (int, error)` - Writes data to buffer

### Query

- `AvailableRead() uint64` - Bytes available for reading
- `AvailableWrite() uint64` - Bytes available for writing
- `Size() uint64` - Total buffer capacity

### Utility

- `Reset()` - Clears the buffer (not thread-safe)

## Thread Safety

This implementation is only safe for single producer, single consumer usage. One goroutine should write, and one goroutine should read. Multiple producers or consumers will cause data races.

## License

MIT

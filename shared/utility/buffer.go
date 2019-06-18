package utility

import (
	"bytes"
	"sync"
)

type Buffer struct {
	mutex     sync.Mutex
	wasClosed bool
	buffer    bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{
		wasClosed: false,
	}
}

func (b *Buffer) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.wasClosed = true
	return nil
}

func (b *Buffer) WasClosed() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.wasClosed
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buffer.Read(p)
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buffer.Write(p)
}

package memfs

import (
	"sync"
)

type memFile struct {
	Buffer
	mutex *sync.RWMutex
	name  string
}

// newMemFile creates a Buffer which byte slice is safe from concurrent access,
// the file itself is not thread-safe.
//
// This means multiple files can work safely on the same byte slice,
// but multiple go routines working on the same file may corrupt the internal pointer structure.
func newMemFile(name string, rwMutex *sync.RWMutex, buf *[]byte) *memFile {
	return &memFile{
		Buffer: NewBuffer(buf),
		mutex:  rwMutex,
		name:   name,
	}
}

func (b memFile) Name() string {
	return b.name
}

func (b *memFile) Read(p []byte) (n int, err error) {
	b.mutex.RLock()
	n, err = b.Buffer.Read(p)
	b.mutex.RUnlock()
	return
}

func (b *memFile) Write(p []byte) (n int, err error) {
	b.mutex.Lock()
	n, err = b.Buffer.Write(p)
	b.mutex.Unlock()
	return
}

func (b *memFile) Seek(offset int64, whence int) (n int64, err error) {
	b.mutex.RLock()
	n, err = b.Buffer.Seek(offset, whence)
	b.mutex.RUnlock()
	return
}

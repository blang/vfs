package memfs

import (
	"errors"
	"io"
	"os"
)

// Buffer is a usable block of data similar to a file
type Buffer interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

// MinBufferSize is the minimal initial allocated buffer size
const MinBufferSize = 512

// ErrTooLarge is thrown if it was not possible to enough memory
var ErrTooLarge = errors.New("Volume too large")

// Buf is a Buffer working on a slice of bytes.
type Buf struct {
	buf *[]byte
	ptr int64
}

// NewBuffer creates a new data volume based on a buffer
func NewBuffer(buf *[]byte) *Buf {
	return &Buf{
		buf: buf,
	}
}

// Seek sets the offset for the next Read or Write on the buffer to offset,
// interpreted according to whence:
// 	0 (os.SEEK_SET) means relative to the origin of the file
// 	1 (os.SEEK_CUR) means relative to the current offset
// 	2 (os.SEEK_END) means relative to the end of the file
// It returns the new offset and an error, if any.
func (v *Buf) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case os.SEEK_SET: // Relative to the origin of the file
		abs = offset
	case os.SEEK_CUR: // Relative to the current offset
		abs = int64(v.ptr) + offset
	case os.SEEK_END: // Relative to the end
		abs = int64(len(*v.buf)) + offset
	default:
		return 0, errors.New("Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("Seek: negative position")
	}
	if abs > int64(len(*v.buf)) {
		return 0, errors.New("Seek: too far")
	}
	v.ptr = abs
	return abs, nil
}

// Write writes len(p) byte to the Buffer.
// It returns the number of bytes written and an error if any.
// Write returns non-nil error when n!=len(p).
func (v *Buf) Write(p []byte) (int, error) {
	l := len(p)
	err := v.grow(l)
	if err != nil {
		return 0, err
	}
	copy((*v.buf)[v.ptr:], p)
	v.ptr += int64(l)
	return l, nil
}

// Close the buffer. Currently no effect.
func (v *Buf) Close() error {
	return nil
}

// Read reads len(p) byte from the Buffer starting at the current offset.
// It returns the number of bytes read and an error if any.
// Returns io.EOF error if pointer is at the end of the Buffer.
func (v *Buf) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if v.ptr >= int64(len(*v.buf)) {
		return 0, io.EOF
	}

	n = copy(p, (*v.buf)[v.ptr:])
	v.ptr += int64(n)
	return
}

func (v *Buf) grow(n int) error {
	m := len(*v.buf)
	if (m + n) > cap(*v.buf) {
		size := 2*cap(*v.buf) + MinBufferSize
		if size < n {
			size = n + MinBufferSize
		}
		buf, err := makeSlice(size)
		if err != nil {
			return err
		}
		copy(buf, *v.buf)
		*v.buf = buf
	}
	*v.buf = (*v.buf)[0 : m+n]
	return nil
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) (b []byte, err error) {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			b = nil
			err = ErrTooLarge
			return
		}
	}()
	b = make([]byte, n)
	return
}

package memfs

import (
	"errors"
	"os"

	"io"
)

type Buffer interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

const MinBufferSize = 512

var (
	ErrTooLarge = errors.New("Volume too large")
)

type buffer struct {
	buf *[]byte
	ptr int64
}

// NewVolume creates a new data volume based on a buffer
func NewBuffer(buf *[]byte) Buffer {
	return &buffer{
		buf: buf,
	}
}

func (v *buffer) Seek(offset int64, whence int) (int64, error) {
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

func (v *buffer) Write(p []byte) (int, error) {
	l := len(p)
	err := v.grow(l)
	if err != nil {
		return 0, err
	}
	copy((*v.buf)[v.ptr:], p)
	v.ptr += int64(l)
	return l, nil
}

// TODO: Change?
func (v *buffer) Close() error {
	return nil
}

func (v *buffer) Read(p []byte) (n int, err error) {
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

func (v *buffer) grow(n int) error {
	m := len(*v.buf)
	if (m + n) > cap(*v.buf) {
		buf, err := makeSlice(2*cap(*v.buf) + MinBufferSize)
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

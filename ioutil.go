package vfs

import (
	"io"
	"os"
)

// WriteFile writes data to a file named by filename on the given Filesystem. If
// the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
//
// This is a port of the stdlib ioutil.WriteFile function.
func WriteFile(fs Filesystem, filename string, data []byte, perm os.FileMode) error {
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

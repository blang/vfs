package vfs

import (
	"os"
)

// Dummy creates a new dummy filesystem which returns the given error on every operation.
func Dummy(err error) Filesystem {
	return &dummyFS{err}
}

type dummyFS struct {
	err error
}

// Create returns dummy error
func (fs dummyFS) Create(name string) (File, error) {
	return nil, fs.err
}

// OpenFile returns dummy error
func (fs dummyFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return nil, fs.err
}

// Remove returns dummy error
func (fs dummyFS) Remove(name string) error {
	return fs.err
}

// Rename returns dummy error
func (fs dummyFS) Rename(oldpath, newpath string) error {
	return fs.err
}

// Mkdir returns dummy error
func (fs dummyFS) Mkdir(name string, perm os.FileMode) error {
	return fs.err
}

// Stat returns dummy error
func (fs dummyFS) Stat(name string) (os.FileInfo, error) {
	return nil, fs.err
}

// Lstat returns dummy error
func (fs dummyFS) Lstat(name string) (os.FileInfo, error) {
	return nil, fs.err
}

// ReadDir returns dummy error
func (fs dummyFS) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, fs.err
}

// DummyFile mocks a File returning an error on every operation
// To create a DummyFS returning a dummyFile instead of an error
// you can your own DummyFS:
//
// 	type writeDummyFS struct {
// 		Filesystem
// 	}
//
// 	func (fs writeDummyFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
// 		return DummyFile(dummyError), nil
// 	}
func DummyFile(err error) File {
	return &dummyFile{err}
}

type dummyFile struct {
	err error
}

// Name returns "dummy"
func (f dummyFile) Name() string {
	return "dummy"
}

// Close returns dummy error
func (f dummyFile) Close() error {
	return f.err
}

// Write returns dummy error
func (f dummyFile) Write(p []byte) (n int, err error) {
	return 0, f.err
}

// Read returns dummy error
func (f dummyFile) Read(p []byte) (n int, err error) {
	return 0, f.err
}

// Seek returns dummy error
func (f dummyFile) Seek(offset int64, whence int) (int64, error) {
	return 0, f.err
}

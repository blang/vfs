package vfs

import (
	"os"
)

// Dummy creates a new dummy filesystem which returns the given error on every operation.
func Dummy(err error) *DummyFS {
	return &DummyFS{err}
}

// DummyFS is dummy filesystem which returns an error on every operation.
// It can be used to mock a full filesystem for testing or fs creation.
type DummyFS struct {
	err error
}

// Create returns dummy error
func (fs DummyFS) Create(name string) (File, error) {
	return nil, fs.err
}

// OpenFile returns dummy error
func (fs DummyFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return nil, fs.err
}

// Remove returns dummy error
func (fs DummyFS) Remove(name string) error {
	return fs.err
}

// Rename returns dummy error
func (fs DummyFS) Rename(oldpath, newpath string) error {
	return fs.err
}

// Mkdir returns dummy error
func (fs DummyFS) Mkdir(name string, perm os.FileMode) error {
	return fs.err
}

// Stat returns dummy error
func (fs DummyFS) Stat(name string) (os.FileInfo, error) {
	return nil, fs.err
}

// Lstat returns dummy error
func (fs DummyFS) Lstat(name string) (os.FileInfo, error) {
	return nil, fs.err
}

// ReadDir returns dummy error
func (fs DummyFS) ReadDir(path string) ([]os.FileInfo, error) {
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
func DummyFile(err error) *DumFile {
	return &DumFile{err}
}

// DumFile represents a dummy File
type DumFile struct {
	err error
}

// Name returns "dummy"
func (f DumFile) Name() string {
	return "dummy"
}

// Close returns dummy error
func (f DumFile) Close() error {
	return f.err
}

// Write returns dummy error
func (f DumFile) Write(p []byte) (n int, err error) {
	return 0, f.err
}

// Read returns dummy error
func (f DumFile) Read(p []byte) (n int, err error) {
	return 0, f.err
}

// Seek returns dummy error
func (f DumFile) Seek(offset int64, whence int) (int64, error) {
	return 0, f.err
}

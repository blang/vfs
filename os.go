package vfs

import (
	"io/ioutil"
	"os"
)

type osFS struct{}

// OS returns a filesystem backed by the filesystem of the os. It wraps os.* stdlib operations.
func OS() Filesystem {
	return &osFS{}
}

// Create wraps os.Create
func (fs osFS) Create(name string) (File, error) {
	return os.Create(name)
}

// OpenFile wraps os.OpenFile
func (fs osFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

// Remove wraps os.Remove
func (fs osFS) Remove(name string) error {
	return os.Remove(name)
}

// Mkdir wraps os.Mkdir
func (fs osFS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

// Rename wraps os.Rename
func (fs osFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Stat wraps os.Stat
func (fs osFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Lstat wraps os.Lstat
func (fs osFS) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

// ReadDir wraps ioutil.ReadDir
func (fs osFS) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

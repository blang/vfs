package vfs

import (
	"io"
	"os"
)

// Filesystem represents an abstract filesystem
type Filesystem interface {
	Create(name string) (File, error)
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	Remove(name string) error
	// RemoveAll(path string) error
	Rename(oldpath, newpath string) error
	Mkdir(name string, perm os.FileMode) error
	// Symlink(oldname, newname string) error
	// TempDir() string
	// Chmod(name string, mode FileMode) error
	// Chown(name string, uid, gid int) error
	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	ReadDir(path string) ([]os.FileInfo, error)
}

// File represents a File with common operations.
// It differs from os.File so e.g. Stat() needs to be called from the Filesystem instead.
//   osfile.Stat() -> filesystem.Stat(file.Name())
type File interface {
	Name() string
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

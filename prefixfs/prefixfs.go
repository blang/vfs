package prefixfs

import (
	"os"

	"github.com/blang/vfs"
)

type prefixFS struct {
	r vfs.Filesystem
	p string
}

// Create returns a file system that prefixes all paths and forwards to root.
func Create(root vfs.Filesystem, prefix string) vfs.Filesystem {
	return prefixFS{root, prefix}
}

func (fs prefixFS) prefix(path string) string {
	return fs.p + string(fs.PathSeparator()) + path
}

func (fs prefixFS) PathSeparator() uint8 { return fs.r.PathSeparator() }

func (fs prefixFS) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	return fs.r.OpenFile(fs.prefix(name), flag, perm)
}

func (fs prefixFS) Remove(name string) error {
	return fs.r.Remove(fs.prefix(name))
}

func (fs prefixFS) Rename(oldpath, newpath string) error {
	return fs.r.Rename(fs.prefix(oldpath), fs.prefix(newpath))
}

func (fs prefixFS) Mkdir(name string, perm os.FileMode) error {
	return fs.r.Mkdir(fs.prefix(name), perm)
}

func (fs prefixFS) Stat(name string) (os.FileInfo, error) {
	return fs.r.Stat(fs.prefix(name))
}

func (fs prefixFS) Lstat(name string) (os.FileInfo, error) {
	return fs.r.Lstat(fs.prefix(name))
}

func (fs prefixFS) ReadDir(path string) ([]os.FileInfo, error) {
	return fs.r.ReadDir(fs.prefix(path))
}

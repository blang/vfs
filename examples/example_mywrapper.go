package examples

import (
	"errors"
	"os"

	"github.com/blang/vfs"
)

type noNewDirs struct {
	vfs.Filesystem
}

func NoNewDirs(fs vfs.Filesystem) vfs.Filesystem {
	return &noNewDirs{fs}
}

// Mkdir is disabled
func (fs *noNewDirs) Mkdir(name string, perm os.FileMode) error {
	return errors.New("Mkdir disabled!")
}

func ExampleMyWrapper() {

	// Disable Mkdirs on the OS Filesystem
	var fs vfs.Filesystem = NoNewDirs(vfs.OS())

	err := fs.Mkdir("/tmp", 0777)
	if err != nil {
		fatal("Mkdir disabled!")
	}
}

package examples

import (
	"errors"

	"github.com/blang/vfs"
)

type myFS struct {
	vfs.Filesystem // Embed the Filesystem interface and fill it with vfs.Dummy on creation
}

func MyFS() *myFS {
	return &myFS{
		vfs.Dummy(errors.New("Not implemented yet!")),
	}
}

func ExampleMyFilesystem() {
	// Simply bootstrap your filesystem
	var fs vfs.Filesystem = MyFS()

	err := fs.Mkdir("/tmp", 0777)
	if err != nil {
		fatal("Error will be: Not implemented yet!")
	}
}

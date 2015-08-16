package examples

import "github.com/blang/vfs"

func ExampleBasicOS() {

	// Create a vfs accessing the filesystem of the underlying OS
	osFS := vfs.OS()
	err := osFS.Mkdir("/tmp/vfs", 0777)
	if err != nil {
		fatal("Error creating directory: %s\n", err)
	}

	f, err := osFS.Create("/tmp/vfs/example.txt")
	if err != nil {
		fatal("Could not create file: %s\n", err)
	}
	defer f.Close()
	if _, err := f.Write([]byte("VFS working on your filesystem")); err != nil {
		fatal("Error writing to file: %s\n", err)
	}
}

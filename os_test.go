package vfs

import (
	"testing"
)

func TestOSCreate(t *testing.T) {
	fs := OS()

	f, err := fs.Create("/tmp/test123")
	if err != nil {
		t.Errorf("Create: %s", err)
	}
	err = f.Close()
	if err != nil {
		t.Errorf("Close: %s", err)
	}
	err = fs.Remove(f.Name())
	if err != nil {
		t.Errorf("Remove: %s", err)
	}
}

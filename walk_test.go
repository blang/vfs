package vfs

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type Node struct {
	name    string
	entries []*Node // nil if the entry is a file
	mark    int
}

var tree = &Node{
	"testdata",
	[]*Node{
		{"a", nil, 0},
		{"b", []*Node{}, 0},
		{"c", nil, 0},
		{
			"d",
			[]*Node{
				{"x", nil, 0},
				{"y", []*Node{}, 0},
				{
					"z",
					[]*Node{
						{"u", nil, 0},
						{"v", nil, 0},
					},
					0,
				},
			},
			0,
		},
	},
	0,
}

func walkTree(n *Node, path string, f func(path string, n *Node)) {
	f(path, n)
	for _, e := range n.entries {
		walkTree(e, filepath.Join(path, e.name), f)
	}
}

func makeTree(t *testing.T, fs Filesystem) {
	walkTree(tree, tree.name, func(path string, n *Node) {
		if n.entries == nil {
			fd, err := fs.OpenFile(path, os.O_CREATE, os.ModePerm)
			if err != nil {
				t.Errorf("makeTree: %v", err)
				return
			}
			fd.Close()
		} else {
			fs.Mkdir(path, 0770)
		}
	})
}

func markTree(n *Node) { walkTree(n, "", func(path string, n *Node) { n.mark++ }) }

func checkMarks(t *testing.T, report bool) {
	walkTree(tree, tree.name, func(path string, n *Node) {
		if n.mark != 1 && report {
			t.Errorf("node %s mark = %d; expected 1", path, n.mark)
		}
		n.mark = 0
	})
}

// Assumes that each node name is unique. Good enough for a test.
// If clear is true, any incoming error is cleared before return. The errors
// are always accumulated, though.
func mark(info os.FileInfo, err error, errors *[]error, clear bool) error {
	name := info.Name()
	walkTree(tree, tree.name, func(path string, n *Node) {
		if n.name == name {
			n.mark++
		}
	})
	if err != nil {
		*errors = append(*errors, err)
		if clear {
			return nil
		}
		return err
	}
	return nil
}

func chtmpdir(t *testing.T) (restore func()) {
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	d, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	return func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("chtmpdir: %v", err)
		}
		os.RemoveAll(d)
	}
}

func TestWalk(t *testing.T) {
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm", "arm64":
			restore := chtmpdir(t)
			defer restore()
		}
	}

	tmpDir, err := ioutil.TempDir("", "TestWalk")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("finding working dir:", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatal("entering temp dir:", err)
	}
	defer os.Chdir(origDir)

	fs := OS()

	makeTree(t, fs)
	errors := make([]error, 0, 10)
	clear := true
	markFn := func(path string, info os.FileInfo, err error) error {
		return mark(info, err, &errors, clear)
	}
	// Expect no errors.
	err = Walk(fs, tree.name, markFn)
	if err != nil {
		t.Fatalf("no error expected, found: %s", err)
	}
	if len(errors) != 0 {
		t.Fatalf("unexpected errors: %s", errors)
	}
	checkMarks(t, true)
	errors = errors[0:0]

	// Test permission errors. Only possible if we're not root
	// and only on some file systems (AFS, FAT).  To avoid errors during
	// all.bash on those file systems, skip during go test -short.
	if os.Getuid() > 0 && !testing.Short() {
		// introduce 2 errors: chmod top-level directories to 0
		os.Chmod(filepath.Join(tree.name, tree.entries[1].name), 0)
		os.Chmod(filepath.Join(tree.name, tree.entries[3].name), 0)

		// 3) capture errors, expect two.
		// mark respective subtrees manually
		markTree(tree.entries[1])
		markTree(tree.entries[3])
		// correct double-marking of directory itself
		tree.entries[1].mark--
		tree.entries[3].mark--
		err := Walk(fs, tree.name, markFn)
		if err != nil {
			t.Fatalf("expected no error return from Walk, got %s", err)
		}
		if len(errors) != 2 {
			t.Errorf("expected 2 errors, got %d: %s", len(errors), errors)
		}
		// the inaccessible subtrees were marked manually
		checkMarks(t, true)
		errors = errors[0:0]

		// 4) capture errors, stop after first error.
		// mark respective subtrees manually
		markTree(tree.entries[1])
		markTree(tree.entries[3])
		// correct double-marking of directory itself
		tree.entries[1].mark--
		tree.entries[3].mark--
		clear = false // error will stop processing
		err = Walk(fs, tree.name, markFn)
		if err == nil {
			t.Fatalf("expected error return from Walk")
		}
		if len(errors) != 1 {
			t.Errorf("expected 1 error, got %d: %s", len(errors), errors)
		}
		// the inaccessible subtrees were marked manually
		checkMarks(t, false)
		errors = errors[0:0]

		// restore permissions
		os.Chmod(filepath.Join(tree.name, tree.entries[1].name), 0770)
		os.Chmod(filepath.Join(tree.name, tree.entries[3].name), 0770)
	}
}

func touch(t *testing.T, fs Filesystem, name string) {
	f, err := fs.OpenFile(name, os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWalkSkipDirOnFile(t *testing.T) {
	td, err := ioutil.TempDir("", "walktest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	if err := os.MkdirAll(filepath.Join(td, "dir"), 0755); err != nil {
		t.Fatal(err)
	}

	fs := OS()

	touch(t, fs, filepath.Join(td, "dir/foo1"))
	touch(t, fs, filepath.Join(td, "dir/foo2"))

	sawFoo2 := false
	walker := func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "foo2") {
			sawFoo2 = true
		}
		if strings.HasSuffix(path, "foo1") {
			return filepath.SkipDir
		}
		return nil
	}

	err = Walk(fs, td, walker)
	if err != nil {
		t.Fatal(err)
	}
	if sawFoo2 {
		t.Errorf("SkipDir on file foo1 did not block processing of foo2")
	}

	err = Walk(fs, filepath.Join(td, "dir"), walker)
	if err != nil {
		t.Fatal(err)
	}
	if sawFoo2 {
		t.Errorf("SkipDir on file foo1 did not block processing of foo2")
	}
}

type statWrapper struct {
	Filesystem

	statErr error
}

func (s *statWrapper) Lstat(path string) (os.FileInfo, error) {
	if strings.HasSuffix(path, "stat-error") {
		return nil, s.statErr
	}
	return os.Lstat(path)
}

func TestWalkFileError(t *testing.T) {
	td, err := ioutil.TempDir("", "walktest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	fs := Filesystem(OS())

	touch(t, fs, filepath.Join(td, "foo"))
	touch(t, fs, filepath.Join(td, "bar"))
	dir := filepath.Join(td, "dir")
	if err := MkdirAll(fs, filepath.Join(td, "dir"), 0755); err != nil {
		t.Fatal(err)
	}
	touch(t, fs, filepath.Join(dir, "baz"))
	touch(t, fs, filepath.Join(dir, "stat-error"))
	statErr := errors.New("some stat error")

	fs = &statWrapper{Filesystem: fs, statErr: statErr}

	got := map[string]error{}
	err = Walk(fs, td, func(path string, fi os.FileInfo, err error) error {
		rel, _ := filepath.Rel(td, path)
		got[filepath.ToSlash(rel)] = err
		return nil
	})
	if err != nil {
		t.Errorf("Walk error: %v", err)
	}
	want := map[string]error{
		".":              nil,
		"foo":            nil,
		"bar":            nil,
		"dir":            nil,
		"dir/baz":        nil,
		"dir/stat-error": statErr,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Walked %#v; want %#v", got, want)
	}
}

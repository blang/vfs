package memfs

import (
	"errors"
	"fmt"
	"os"
	filepath "path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/blang/vfs"
)

var (
	ErrReadOnly  = errors.New("File is read-only")
	ErrWriteOnly = errors.New("File is write-only")
)

const PathSeparator = "/"

type memFS struct {
	root *fileInfo
	wd   *fileInfo
	lock *sync.RWMutex
}

type fileInfo struct {
	name    string
	dir     bool
	mode    os.FileMode
	parent  *fileInfo
	size    int64
	modTime time.Time
	fs      vfs.Filesystem
	childs  map[string]*fileInfo
	buf     *[]byte
	mutex   *sync.RWMutex
}

func (fi fileInfo) Sys() interface{} {
	return fi.fs
}

func (fi fileInfo) Size() int64 {
	fi.mutex.RLock()
	l := len(*(fi.buf))
	fi.mutex.RUnlock()
	return int64(l)
}

func (fi fileInfo) IsDir() bool {
	return fi.dir
}

func (fi fileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi fileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi fileInfo) Name() string {
	return fi.name
}

func (fi fileInfo) AbsPath() string {
	if fi.parent != nil {
		return filepath.Join(fi.parent.AbsPath(), fi.name)
	}
	return "/"
}

// MemFS creates a new filesystem which entirely resides in memory
func MemFS() vfs.Filesystem {
	root := &fileInfo{
		name: "/",
		dir:  true,
	}
	return &memFS{
		root: root,
		wd:   root,
		lock: &sync.RWMutex{},
	}
}

// Mkdir creates a new directory with given permissions
func (fs *memFS) Mkdir(name string, perm os.FileMode) error {
	name = filepath.Clean(name)
	dirPath, base := filepath.Split(name)
	parent, err := fs.findFileinfo(dirPath)
	if err != nil {
		return &os.PathError{"mkdir", name, err}
	}

	if parent.childs == nil {
		parent.childs = make(map[string]*fileInfo)
	} else {
		// Check if dir already exists
		if _, ok := parent.childs[base]; ok {
			return &os.PathError{"mkdir", name, fmt.Errorf("Directory %q already exists", name)}
		}
	}

	fi := &fileInfo{
		name:   base,
		dir:    true,
		mode:   perm,
		parent: parent,
		fs:     fs,
	}
	parent.childs[base] = fi
	return nil
}

// byName implements sort.Interface
type byName []os.FileInfo

// Len returns the length of the slice
func (f byName) Len() int { return len(f) }

// Less sorts slice by Name
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }

// Swap two elements by index
func (f byName) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func (fs *memFS) ReadDir(path string) ([]os.FileInfo, error) {
	path = filepath.Clean(path)
	fi, err := fs.findFileinfo(path)
	if err != nil {
		return nil, &os.PathError{"readdir", path, err}
	}
	if !fi.dir {
		return nil, &os.PathError{"readdir", path, errors.New("Not a directory")}
	}

	fis := make([]os.FileInfo, 0, len(fi.childs))
	for _, e := range fi.childs {
		fis = append(fis, e)
	}
	sort.Sort(byName(fis))
	return fis, nil
}

// findFileinfo searches the filetree beginning at root and returns the fileInfo of the file at path
func (fs *memFS) findFileinfo(dir string) (*fileInfo, error) {
	dir = filepath.Clean(dir)
	segments := SplitPath(dir, PathSeparator)
	if len(segments) == 1 {
		if segments[0] == "" {
			return fs.root, nil
		} else if segments[0] == "." {
			return fs.wd, nil
		}
	}

	// log.Printf("Dir: %s, Segments: %q (%d)", dir, segments, len(segments))
	parent := fs.root
	if len(segments) > 0 && segments[0] == "." {
		segments = segments[1:]
		parent = fs.wd
	}
	if len(segments) > 0 && strings.TrimSpace(segments[0]) == "" {
		segments = segments[1:]
	}
	// TODO: Could not find files? check parent.dir before every iteration?
	for i, seg := range segments {
		if parent.childs == nil {
			return parent, fmt.Errorf("Directory parent %q does not exist: %q", filepath.Join(segments[:i]...))
		}
		if entry, ok := parent.childs[seg]; ok && entry.dir {
			parent = entry
		} else {
			return parent, fmt.Errorf("Directory parent %q does not exist: %q", filepath.Join(segments[:i]...))
		}
	}
	return parent, nil
}

func (fs *memFS) Create(name string) (vfs.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func checkFlag(flag int, flags int) bool {
	return flags&flag == flag
}

func (fs *memFS) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	name = filepath.Clean(name)
	dir, base := filepath.Split(name)
	fiParent, err := fs.findFileinfo(dir)
	if err != nil {
		return nil, err
	}

	var fi *fileInfo
	if checkFlag(os.O_CREATE, flag) {
		if fiParent.childs == nil {
			fiParent.childs = make(map[string]*fileInfo)
		} else {
			if _, ok := fiParent.childs[base]; ok {

				// If O_TRUNC is set, existing file is overwritten
				if !checkFlag(os.O_TRUNC, flag) {
					return nil, os.ErrExist
				}
			}
		}
		fi = &fileInfo{
			name:   base,
			dir:    false,
			mode:   perm,
			parent: fiParent,
			fs:     fs,
		}
		fiParent.childs[base] = fi
	} else { // find existing
		if fiParent.childs == nil {
			return nil, os.ErrNotExist
		}
		var ok bool
		fi, ok = fiParent.childs[base]
		if !ok {
			return nil, os.ErrNotExist
		}

	}

	return fi.File(flag)
}

func (fi *fileInfo) File(flag int) (vfs.File, error) {
	if fi.buf == nil || checkFlag(os.O_TRUNC, flag) {
		buf := make([]byte, 0, MinBufferSize)
		fi.buf = &buf
	}
	if fi.mutex == nil {
		fi.mutex = &sync.RWMutex{}
	}
	var f vfs.File = newMemFile(fi.AbsPath(), fi.mutex, fi.buf)
	if checkFlag(os.O_APPEND, flag) {
		f.Seek(0, os.SEEK_END)
	}
	if checkFlag(os.O_RDWR, flag) {
		return f, nil
	} else if checkFlag(os.O_WRONLY, flag) {
		f = &woFile{f}
	} else {
		f = &roFile{f}
	}

	return f, nil
}

// roFile wraps the given file and disables Write(..) operation.
type roFile struct {
	vfs.File
}

// Write is disabled and returns ErrorReadOnly
func (f *roFile) Write(p []byte) (n int, err error) {
	return 0, ErrReadOnly
}

// woFile wraps the given file and disables Read(..) operation.
type woFile struct {
	vfs.File
}

// Read is disabled and returns ErrorWroteOnly
func (f *woFile) Read(p []byte) (n int, err error) {
	return 0, ErrWriteOnly
}

func (fs *memFS) Remove(name string) error {
	name = filepath.Clean(name)
	dir, base := filepath.Split(name)
	fiParent, err := fs.findFileinfo(dir)
	if err != nil {
		return &os.PathError{"remove", name, err}
	}
	if _, ok := fiParent.childs[base]; ok {
		delete(fiParent.childs, base)
		return nil
	}
	return &os.PathError{"remove", name, os.ErrNotExist}
}

func (fs *memFS) Rename(oldpath, newpath string) error {
	// OldPath
	oldpath = filepath.Clean(oldpath)
	oldDir, oldBase := filepath.Split(oldpath)
	fiOldParent, err := fs.findFileinfo(oldDir)
	if err != nil {
		return &os.PathError{"rename", oldpath, err}
	}
	fiOld, ok := fiOldParent.childs[oldBase]
	if !ok {
		return &os.PathError{"rename", oldpath, os.ErrNotExist}
	}

	newpath = filepath.Clean(newpath)
	newDir, newBase := filepath.Split(newpath)
	fiNewParent, err := fs.findFileinfo(newDir)
	if err != nil {
		return &os.PathError{"rename", newpath, err}
	}

	if fiNewParent.childs == nil {
		fiNewParent.childs = make(map[string]*fileInfo)
	}

	if _, ok := fiNewParent.childs[newBase]; ok {
		return &os.PathError{"rename", newpath, os.ErrExist}
	}

	delete(fiOldParent.childs, oldBase)
	fiOld.parent = fiNewParent
	fiNewParent.childs[newBase] = fiOld
	return nil
}

func (fs *memFS) Stat(name string) (os.FileInfo, error) {
	name = filepath.Clean(name)
	dir, base := filepath.Split(name)
	fiParent, err := fs.findFileinfo(dir)
	if err != nil {
		return nil, &os.PathError{"stat", name, err}
	}
	if fi, ok := fiParent.childs[base]; ok {
		return fi, nil
	}
	return nil, &os.PathError{"stat", name, os.ErrNotExist}
}

func (fs *memFS) Lstat(name string) (os.FileInfo, error) {
	return fs.Stat(name)
}

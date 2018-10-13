package vfs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Walk walks the file tree rooted at root, calling walkFunc for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func Walk(fs Filesystem, root string, walkFunc filepath.WalkFunc) error {
	info, err := fs.Lstat(root)
	if err != nil {
		err = walkFunc(root, nil, err)
	} else {
		err = walk(fs, root, info, walkFunc)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(fs Filesystem, dirname string) ([]string, error) {
	infos, err := fs.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name())
	}
	sort.Strings(names)
	return names, nil
}

// walk recursively descends path, calling walkFunc.
func walk(fs Filesystem, path string, info os.FileInfo, walkFunc filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFunc(path, info, nil)
	}

	names, err := readDirNames(fs, path)
	err1 := walkFunc(path, info, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means walkFn want walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err != nil || err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by walkFn. walkFn may ignore err and return nil.
		// If walkFn returns SkipDir, it will be handled by the caller.
		// So walk should return whatever walkFn returns.
		return err1
	}

	for _, name := range names {
		filename := strings.Join([]string{path, name}, string(fs.PathSeparator()))
		fileInfo, err := fs.Lstat(filename)
		if err != nil {
			if err := walkFunc(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(fs, filename, fileInfo, walkFunc)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

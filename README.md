vfs for golang [![Build Status](https://drone.io/github.com/blang/vfs/status.png)](https://drone.io/github.com/blang/vfs/latest) [![GoDoc](https://godoc.org/github.com/blang/vfs?status.png)](https://godoc.org/github.com/blang/vfs) [![Coverage Status](https://img.shields.io/coveralls/blang/vfs.svg)](https://coveralls.io/r/blang/vfs?branch=master) [![Join the chat at https://gitter.im/blang/vfs](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/blang/vfs?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
======

vfs is library to support virtual filesystems. It provides basic abstractions of filesystems and implementations, like `OS` accessing the file system of the underlying OS and `memfs` a full filesystem in-memory.

Usage
-----
```bash
$ go get github.com/blang/vfs
```
Note: Always vendor your dependencies or fix on a specific version tag.

```go
import github.com/blang/vfs
```

```go
// Create a vfs accessing the filesystem of the underlying OS
fs := vfs.OS()
fs.Mkdir("/tmp", 0777)

// Make the filesystem read-only:
fs = vfs.ReadOnly(fs) // Simply wrap filesystems to change its behaviour

// os.O_CREATE will fail and return vfs.ErrReadOnly
// os.O_RDWR is supported but Write(..) on the file is disabled
f, _ := fs.OpenFile("/tmp/example.txt", os.O_RDWR, 0)

// Return vfs.ErrReadOnly
_, err := f.Write([]byte("Write on readonly fs?"))

// Create a fully writable filesystem in memory
fs := memfs.Create()
fs.Mkdir("/root")
```

Check detailed examples below. Also check the [GoDocs](http://godoc.org/github.com/blang/vfs).

Why should I use this lib?
-----

- Only Stdlib
- (Nearly) Fully tested (Coverage >90%)
- Easy to create your own filesystem
- Mock a full filesystem for testing (or use included `memfs`)
- Compose/Wrap Filesystems `ReadOnly(OS())` and write simple Wrappers
- Many features, see [GoDocs](http://godoc.org/github.com/blang/vfs) and examples below

Example
-----

Have a look at full examples in [examples/](examples/)

Features
-----

- OS Filesystem support
- ReadOnly Wrapper 
- DummyFS for quick mocking
- MemFS full in-memory filesystem
- MountFS - support mounts across filesystems

Current state: ALPHA
-----

While the functionality is quite stable and heavily tested, interfaces are subject to change. 

    You need more/less abstraction? Let me know by creating a Issue, thank you.

Motivation
-----

I simply couldn't find any lib supporting this wide range of variation and adaptability.

Contribution
-----

Feel free to make a pull request. For bigger changes create a issue first to discuss about it.

License
-----

See [LICENSE](LICENSE) file.
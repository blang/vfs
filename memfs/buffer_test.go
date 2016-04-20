package memfs

import (
	"os"
	"strings"

	"testing"
)

const (
	dots = "1....2....3....4"
	abc  = "abcdefghijklmnop"
)

var (
	large = strings.Repeat("0123456789", 200) // 2000 bytes
)

func TestWrite(t *testing.T) {
	buf := make([]byte, 0, len(dots))
	v := NewBuffer(&buf)

	// Write first dots
	if n, err := v.Write([]byte(dots)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(dots) {
		t.Errorf("Invalid write count: %d", n)
	}
	if s := string(buf[:len(dots)]); s != dots {
		t.Errorf("Invalid buffer content: %q", s)
	}

	// Write second time: abc - buffer must grow
	if n, err := v.Write([]byte(abc)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(abc) {
		t.Errorf("Invalid write count: %d", n)
	}
	if s := string(buf[:len(dots+abc)]); s != dots+abc {
		t.Errorf("Invalid buffer content: %q", s)
	}

	if len(buf) != len(dots)+len(abc) {
		t.Errorf("Origin buffer did not grow: len=%d, cap=%d", len(buf), cap(buf))
	}

	// Test on case when no buffer grow is needed
	if n, err := v.Seek(0, os.SEEK_SET); err != nil || n != 0 {
		t.Errorf("Invalid seek result: %d %s", n, err)
	}

	// Write dots on start of the buffer
	if n, err := v.Write([]byte(dots)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(dots) {
		t.Errorf("Invalid write count: %d", n)
	}
	if s := string(buf[:len(dots)]); s != dots {
		t.Errorf("Invalid buffer content: %q", s)
	}

	if len(buf) != len(dots)+len(abc) {
		t.Errorf("Origin buffer should not grow: len=%d, cap=%d", len(buf), cap(buf))
	}

	// Restore seek cursor
	if n, err := v.Seek(0, os.SEEK_END); err != nil {
		t.Errorf("Invalid seek result: %d %s", n, err)
	}

	// Can not read, ptr at the end
	p := make([]byte, len(dots))
	if n, err := v.Read(p); err == nil || n > 0 {
		t.Errorf("Expected read error: %d %s", n, err)
	}

	if n, err := v.Seek(0, os.SEEK_SET); err != nil || n != 0 {
		t.Errorf("Invalid seek result: %d %s", n, err)
	}

	// Read dots
	if n, err := v.Read(p); err != nil || n != len(dots) || string(p) != dots {
		t.Errorf("Unexpected read error: %d %s, res: %s", n, err, string(p))
	}

	// Read abc
	if n, err := v.Read(p); err != nil || n != len(abc) || string(p) != abc {
		t.Errorf("Unexpected read error: %d %s, res: %s", n, err, string(p))
	}

	// Seek abc backwards
	if n, err := v.Seek(int64(-len(abc)), os.SEEK_END); err != nil || n != int64(len(dots)) {
		t.Errorf("Invalid seek result: %d %s", n, err)
	}

	// Seek 8 forwards
	if n, err := v.Seek(int64(len(abc)/2), os.SEEK_CUR); err != nil || n != int64(16)+int64(len(abc)/2) {
		t.Errorf("Invalid seek result: %d %s", n, err)
	}

	// Seek to end
	if n, err := v.Seek(0, os.SEEK_END); err != nil || n != int64(len(buf)) {
		t.Errorf("Invalid seek result: %d %s", n, err)
	}

	// Write so that buffer must expand more than 2x
	if n, err := v.Write([]byte(large)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(large) {
		t.Errorf("Invalid write count: %d", n)
	}
	if s := string(buf[:len(dots+abc+large)]); s != dots+abc+large {
		t.Errorf("Invalid buffer content: %q", s)
	}

	if len(buf) != len(dots)+len(abc)+len(large) {
		t.Errorf("Origin buffer did not grow: len=%d, cap=%d", len(buf), cap(buf))
	}
}

func TestVolumesConcurrentAccess(t *testing.T) {
	buf := make([]byte, 0, len(dots))
	v1 := NewBuffer(&buf)
	v2 := NewBuffer(&buf)

	// v1 write dots
	if n, err := v1.Write([]byte(dots)); err != nil || n != len(dots) {
		t.Errorf("Unexpected write error: %d %s", n, err)
	}

	p := make([]byte, len(dots))

	// v2 read dots
	if n, err := v2.Read(p); err != nil || n != len(dots) || string(p) != dots {
		t.Errorf("Unexpected read error: %d %s, res: %s", n, err, string(p))
	}

	// v2 write dots
	if n, err := v2.Write([]byte(abc)); err != nil || n != len(abc) {
		t.Errorf("Unexpected write error: %d %s", n, err)
	}

	// v1 read dots
	if n, err := v1.Read(p); err != nil || n != len(abc) || string(p) != abc {
		t.Errorf("Unexpected read error: %d %s, res: %s", n, err, string(p))
	}

}

func TestSeek(t *testing.T) {
	buf := make([]byte, 0, len(dots))
	v := NewBuffer(&buf)

	// write dots
	if n, err := v.Write([]byte(dots)); err != nil || n != len(dots) {
		t.Errorf("Unexpected write error: %d %s", n, err)
	}

	// invalid whence
	if _, err := v.Seek(0, 4); err == nil {
		t.Errorf("Expected invalid whence error")
	}
	// seek to -1
	if _, err := v.Seek(-1, os.SEEK_SET); err == nil {
		t.Errorf("Expected invalid position error")
	}

	// seek to end
	if _, err := v.Seek(0, os.SEEK_END); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	// seel past the end
	if _, err := v.Seek(1, os.SEEK_END); err == nil {
		t.Errorf("Expected invalid position error")
	}
}

func TestRead(t *testing.T) {
	buf := make([]byte, len(dots))
	copy(buf, []byte(dots))
	v := NewBuffer(&buf)

	// p := make([]byte, len(dots))
	var p []byte

	// Read to empty buffer, err==nil, n == 0
	if n, err := v.Read(p); err != nil || n > 0 {
		t.Errorf("Unexpected read error: %d %s, res: %s", n, err, string(p))
	}
}

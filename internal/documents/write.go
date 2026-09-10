package documents

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileIfChanged atomically replaces one regular file, preserving permissions.
// Identical bytes do not change mtime. This is not a multi-file transaction or
// protection against an uncooperative editor racing between read and rename.
func WriteFileIfChanged(path string, data []byte) (bool, error) {
	mode := os.FileMode(0644)
	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() {
			return false, fmt.Errorf("not a regular file: %s", path)
		}
		mode = info.Mode().Perm()
		old, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}
		if bytes.Equal(old, data) {
			return false, nil
		}
	} else if !os.IsNotExist(err) {
		return false, err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, err
	}
	tmp, err := os.CreateTemp(dir, ".docmgr-*")
	if err != nil {
		return false, err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return false, err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return false, err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return false, err
	}
	if err := tmp.Close(); err != nil {
		return false, err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return false, err
	}
	return true, nil
}

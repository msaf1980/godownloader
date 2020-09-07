package fileutils

import (
	"os"
	"syscall"
)

// Exist check for path exist (exist, isDir, error)
func Exist(path string) (bool, bool, error) {
	if s, err := os.Stat(path); err == nil {
		if s.IsDir() {
			return true, true, nil
		}
		return true, false, nil
	} else if !os.IsNotExist(err) {
		return false, false, nil
	} else {
		pathErr, ok := err.(*os.PathError)
		if ok && pathErr.Err == syscall.ENOENT {
			return false, false, nil
		}
		return false, false, err
	}
}

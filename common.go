package xbps

import (
	"os"
)

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fi.IsRegular() || fi.IsDirectory() {
		return true
	}
	return false
}

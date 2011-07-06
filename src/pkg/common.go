package xbps

import (
	"os"
	"strings"
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

func unExpand(path string) string {
	home := os.Getenv("HOME")
	if strings.Contains(path, home) {
		return strings.Replace(path, home, "~", 1)
	}
	return path
}

package main

import (
	"net/http"
	"os"
)

// fallback opens defaultPath when the underlying fs returns os.ErrNotExist
type fallback struct {
	defaultPath string
	fs          http.FileSystem
}

func (fb fallback) Open(path string) (http.File, error) {
	f, err := fb.fs.Open(path)
	if os.IsNotExist(err) {
		return fb.fs.Open(fb.defaultPath)
	}
	return f, err
}

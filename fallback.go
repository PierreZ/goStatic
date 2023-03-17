package main

import (
	"net/http"
	"os"
	"path"
)

// fallback opens defaultPath when the underlying fs returns os.ErrNotExist
type fallback struct {
	defaultPath string
	fs          http.FileSystem
}

func OpenDefault(fb fallback, requestPath string) (http.File, error) {
	requestPath = path.Dir(requestPath)
	defaultFile := requestPath + "/" + fb.defaultPath

	f, err := fb.fs.Open(defaultFile)
	if os.IsNotExist(err) && requestPath != "" {
		parentPath, _ := path.Split(requestPath)
		return OpenDefault(fb, parentPath)
	}
	return f, err
}

func (fb fallback) Open(requestPath string) (http.File, error) {
	f, err := fb.fs.Open(requestPath)
	if os.IsNotExist(err) {
		if len(fb.defaultPath) == 0 || fb.defaultPath[0] == '/' {
			return fb.fs.Open(fb.defaultPath)
		}
		return OpenDefault(fb, requestPath)
	}
	return f, err
}

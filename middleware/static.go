package middleware

import (
	"net/http"
	"os"
	"strings"
)

// NoDirectoryListingFileSystem adalah wrapper untuk http.FileSystem yang mencegah directory listing
type NoDirectoryListingFileSystem struct {
	fs http.FileSystem
}

func (fs NoDirectoryListingFileSystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	// Jika ini adalah direktori, return error untuk mencegah listing
	if stat.IsDir() {
		f.Close()
		return nil, os.ErrNotExist
	}

	return f, nil
}

// NoDirectoryListingHandler mengembalikan handler yang mencegah directory listing
func NoDirectoryListingHandler(root http.FileSystem, prefix string) http.Handler {
	fileServer := http.FileServer(NoDirectoryListingFileSystem{fs: root})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Type header untuk berbagai file types
		path := r.URL.Path
		if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		} else if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}

		fileServer.ServeHTTP(w, r)
	})
}

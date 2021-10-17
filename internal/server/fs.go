package server

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

func (s *server) fsWrapper(fs http.FileSystem) http.HandlerFunc {
	fserv := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = fmt.Sprintf("/%s", r.URL.Path)
		}

		clean := path.Clean(r.URL.Path)
		f, err := fs.Open(clean)
		if err != nil {
			if os.IsNotExist(err) {
				s.logger.WithField("requestPath", r.URL.Path).Debug("path does not exist, rewriting to /")
				r.URL.Path = "/"
			}
		}

		if err == nil {
			f.Close()
		}

		fserv.ServeHTTP(w, r)
	})
}

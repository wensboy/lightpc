package ui

import (
	"bytes"
	"io/fs"
	"maps"
	"net/http"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// spaFallback answers index.html for a path the asset server could not resolve.
//
// The next handler runs into a buffer first. Deciding from the status afterwards needs the response to still be writable,
// and deciding from the path beforehand is worse: the asset server serves endpoints that carry no file extension, and a
// request for one of those would be swapped for the shell, leaving whatever expected it with HTML.
//
// Only a 404 is replaced, and only when the client asked for a document. A missing script or font keeps its status: a
// browser handed HTML where it expects code reports a syntax error, which buries the actual mistake.
func SpaFallback(fsys fs.FS) application.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buffered := newBufferedWriter(w)
			next.ServeHTTP(buffered, r)

			if buffered.status != http.StatusNotFound || !wantsDocument(r) {
				buffered.flush(w)
				return
			}

			index, err := fs.ReadFile(fsys, "index.html")
			if err != nil {
				buffered.flush(w)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(index)
		})
	}
}

// wantsDocument reports whether the client is after a page rather than a file.
//
// A navigation sends text/html in Accept and a document-like Sec-Fetch-Dest; a script, a stylesheet or an API call sends
// neither, and those must keep the 404 the asset server produced.
func wantsDocument(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		return false
	}
	switch r.Header.Get("Sec-Fetch-Dest") {
	case "", "document", "iframe":
		return true
	default:
		return false
	}
}

// bufferedWriter holds the response until the caller decides whether to keep it.
//
// The status and headers are remembered rather than forwarded, because a header written once cannot be revised. The body
// is kept too: a shell that replaces a 404 replaces its body as well, and the buffered one must not leak through.
type bufferedWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newBufferedWriter(origin http.ResponseWriter) *bufferedWriter {
	return &bufferedWriter{header: origin.Header().Clone(), status: http.StatusOK}
}

// Header implements http.ResponseWriter.
func (b *bufferedWriter) Header() http.Header { return b.header }

// WriteHeader implements http.ResponseWriter: it records the status instead of sending it.
func (b *bufferedWriter) WriteHeader(status int) { b.status = status }

// Write implements http.ResponseWriter.
func (b *bufferedWriter) Write(data []byte) (int, error) { return b.body.Write(data) }

// flush copies the held response onto the origin.
func (b *bufferedWriter) flush(origin http.ResponseWriter) {
	maps.Copy(origin.Header(), b.header)
	origin.WriteHeader(b.status)
	if b.status != http.StatusNoContent && b.status != http.StatusNotModified {
		_, _ = origin.Write(b.body.Bytes())
	}
}

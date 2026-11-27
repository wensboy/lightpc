package ui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

// bundle is a miniature asset tree: the shell and one script.
func bundle() fs.FS {
	return fstest.MapFS{
		"index.html":     {Data: []byte("<html>shell</html>")},
		"assets/app.js":  {Data: []byte("console.log(1)")},
		"assets/app.css": {Data: []byte("body{}")},
	}
}

// navigation is the request a browser sends when it opens or reloads a page.
func navigation(target string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("Sec-Fetch-Dest", "document")
	return request
}

// asset is the request a document sends for something it references.
func asset(target string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Sec-Fetch-Dest", "script")
	return request
}

// serve runs one request through the fallback wrapped around a handler that answers 404 for everything, which is what
// the asset server does for a path it cannot map onto a file.
func serve(t *testing.T, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	failing := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	recorder := httptest.NewRecorder()
	SpaFallback(bundle())(failing).ServeHTTP(recorder, request)
	return recorder
}

func TestFallbackAnswersANavigationWithTheShell(t *testing.T) {
	// A deep link has no file behind it; the router resolves it once the shell has loaded.
	response := serve(t, navigation("/dashboard/settings"))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if body := response.Body.String(); body != "<html>shell</html>" {
		t.Fatalf("body = %q", body)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Fatalf("content type = %q", contentType)
	}
}

func TestFallbackLeavesAnExtensionlessEndpointAlone(t *testing.T) {
	// The asset server serves endpoints that carry no file extension. Swapping one for the shell would hand whatever
	// expects a payload an HTML document instead, which is how a path-based guess breaks a working runtime.
	for _, target := range []string{"/wails/runtime", "/wails/events", "/assets/bundle"} {
		response := serve(t, asset(target))
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s: status = %d", target, response.Code)
		}
	}
}

func TestFallbackLeavesAMissingAssetAlone(t *testing.T) {
	// A browser handed HTML where it expects code reports a syntax error, burying the actual 404.
	response := serve(t, asset("/assets/missing.js"))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestFallbackPassesThroughWhatTheAssetsHold(t *testing.T) {
	// The asset server answers this in a real application; the handler used here answers 404 for everything, which is
	// how the test tells a pass through from a served shell.
	response := serve(t, asset("/assets/app.js"))

	if response.Code != http.StatusNotFound {
		t.Fatalf("the fallback answered for a file it does not own: status = %d", response.Code)
	}
}

func TestFallbackLeavesNonNavigationRequestsAlone(t *testing.T) {
	// A form post, a fetch, a head against an expensive path: all of them carry a bodiless Accept list, and none of
	// them can do anything with a shell.
	post := httptest.NewRequest(http.MethodPost, "/dashboard", nil)
	post.Header.Set("Accept", "text/html")
	if response := serve(t, post); response.Code != http.StatusNotFound {
		t.Fatalf("post: status = %d", response.Code)
	}

	plain := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	plain.Header.Set("Accept", "application/json")
	if response := serve(t, plain); response.Code != http.StatusNotFound {
		t.Fatalf("json get: status = %d", response.Code)
	}
}

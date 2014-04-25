package loader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPFetcher(t *testing.T) {
	content := `Hello, World!`
	modtime := time.Now().Add(-1 * time.Hour).UTC()
	modtime = modtime.Add(-1 * time.Duration(modtime.Nanosecond()))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hello.tx":
			w.Header().Set("Last-Modified", modtime.Format(http.TimeFormat))
			fmt.Fprintf(w, content)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	}))
	defer ts.Close()

	f, err := NewHTTPTemplateFetcher([]string{ts.URL})
	if err != nil {
		t.Fatalf("failed to instantiate fetcher: %s", err)
	}

	s, err := f.FetchTemplate("hello.tx")
	if err != nil {
		t.Fatalf("failed to fetch template 'hello.tx': %s", err)
	}

	if lastmod, err := s.LastModified(); err != nil || lastmod != modtime {
		t.Errorf("last-modified does not match. got '%s', expected '%s'", lastmod, modtime)
	}

	if b, err := s.Bytes(); err != nil || string(b) != content {
		t.Errorf("content does not match. got '%s'", b)
	}

}

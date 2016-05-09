package loader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// NewHTTPTemplateFetcher creates a new struct. `urls` must give us the
// base HTTP urls for us to look the templates in (note: do not use trailing slashes)
func NewHTTPTemplateFetcher(urls []string) (*HTTPTemplateFetcher, error) {
	f := &HTTPTemplateFetcher{
		URLs: make([]string, len(urls)),
	}
	for k, v := range urls {
		u, err := url.Parse(v)
		if err != nil {
			return nil, err
		}

		if !u.IsAbs() {
			return nil, fmt.Errorf("url %s is not an absolute url", v)
		}
		f.URLs[k] = u.String()
	}
	return f, nil
}

// FetchTemplate returns a TemplateSource representing the template at path
// `path`. Paths are searched relative to the urls given to NewHTTPTemplateFetcher()
func (l *HTTPTemplateFetcher) FetchTemplate(path string) (TemplateSource, error) {
	u, err := url.Parse(path)

	if err != nil {
		return nil, fmt.Errorf("error parsing given path as url: %s", err)
	}

	if u.IsAbs() {
		return nil, ErrAbsolutePathNotAllowed
	}

	// XXX Consider caching!
	for _, base := range l.URLs {
		u := base + "/" + path
		res, err := http.Get(u)
		if err != nil {
			continue
		}

		return NewHTTPSource(res)
	}
	return nil, ErrTemplateNotFound
}

// NewHTTPSource creates a new HTTPSource instance
func NewHTTPSource(r *http.Response) (*HTTPSource, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	s := &HTTPSource{
		bytes.NewBuffer(body),
		time.Time{},
	}

	if lastmodStr := r.Header.Get("Last-Modified"); lastmodStr != "" {
		t, err := time.Parse(http.TimeFormat, lastmodStr)
		if err != nil {
			fmt.Printf("failed to parse: %s\n", err)
			t = time.Now()
		}
		s.LastModifiedTime = t
	} else {
		s.LastModifiedTime = time.Now()
	}

	return s, nil
}

// LastModified returns the last modified date of this template
func (s *HTTPSource) LastModified() (time.Time, error) {
	return s.LastModifiedTime, nil
}

// Reader returns the io.Reader for the template
func (s *HTTPSource) Reader() (io.Reader, error) {
	return s.Buffer, nil
}

// Bytes returns the bytes in the template file
func (s *HTTPSource) Bytes() ([]byte, error) {
	return s.Buffer.Bytes(), nil
}

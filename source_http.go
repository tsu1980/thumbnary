package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type HttpImageSource struct {
	Config *SourceConfig
}

func NewHttpImageSource(config *SourceConfig) ImageSource {
	return &HttpImageSource{config}
}

func (s *HttpImageSource) GetImage(ireq *http.Request, origin *Origin, filePath string, isExternalHTTPSource bool) ([]byte, error) {
	var uri *url.URL
	var err error
	if isExternalHTTPSource {
		urlStr, err := url.QueryUnescape(filePath)
		if err != nil {
			return nil, err
		}
		uri, err = url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
	} else {
		u := &url.URL{
			Scheme: origin.Scheme,
			Host:   origin.Host,
			Path:   path.Join(origin.PathPrefix, filePath),
		}
		uri, err = url.Parse(u.String())
		if err != nil {
			return nil, err
		}
	}

	return s.fetchImage(uri, ireq)
}

func (s *HttpImageSource) fetchImage(url *url.URL, ireq *http.Request) ([]byte, error) {
	// Check remote image size by fetching HTTP Headers
	if s.Config.MaxAllowedSize > 0 {
		req := newHTTPRequest(s, ireq, "HEAD", url)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Error fetching image http headers: %v", err)
		}
		res.Body.Close()
		if res.StatusCode < 200 && res.StatusCode > 206 {
			return nil, fmt.Errorf("Error fetching image http headers: (status=%d) (url=%s)", res.StatusCode, req.URL.String())
		}

		contentLength, _ := strconv.Atoi(res.Header.Get("Content-Length"))
		if contentLength > s.Config.MaxAllowedSize {
			return nil, fmt.Errorf("Content-Length %d exceeds maximum allowed %d bytes", contentLength, s.Config.MaxAllowedSize)
		}
	}

	// Perform the request using the default client
	req := newHTTPRequest(s, ireq, "GET", url)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error downloading image: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Error downloading image: (status=%d) (url=%s)", res.StatusCode, req.URL.String())
	}

	// Read the body
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to create image from response body: %s (url=%s)", req.URL.String(), err)
	}
	return buf, nil
}

func (s *HttpImageSource) setAuthorizationHeader(req *http.Request, ireq *http.Request) {
	auth := s.Config.Authorization
	if auth == "" {
		auth = ireq.Header.Get("X-Forward-Authorization")
	}
	if auth == "" {
		auth = ireq.Header.Get("Authorization")
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
}

func newHTTPRequest(s *HttpImageSource, ireq *http.Request, method string, url *url.URL) *http.Request {
	req, _ := http.NewRequest(method, url.String(), nil)
	req.Header.Set("User-Agent", "thumbnary/"+Version)
	req.URL = url

	// Forward auth header to the target server, if necessary
	if s.Config.AuthForwarding || s.Config.Authorization != "" {
		s.setAuthorizationHeader(req, ireq)
	}

	return req
}

func init() {
	RegisterSource(ImageSourceTypeHttp, NewHttpImageSource)
}

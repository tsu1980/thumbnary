package main

import (
	"net/http"
	"testing"
)

func TestOriginSlugDetect_Host(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectHostPattern: `([a-z0-9]+)\.example\.test`,
	}
	var originSlugExpected OriginSlug = "klj8a"
	req, _ := http.NewRequest("GET", "http://klj8a.example.test/", nil)
	req.Host = req.URL.Host
	imgReq := &ImageRequest{
		HTTPRequest: req,
	}

	originSlug, err := OriginSlugDetectFunc_Host(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin slug from host name: %s. (req=%+v)", err.Error(), req)
	}

	if originSlug != originSlugExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originSlugExpected, originSlug, req)
	}
}

func TestOriginSlugDetect_Path(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectPathPattern: `^/([a-z0-9]+)/c!/`,
	}
	var originSlugExpected OriginSlug = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/klj8a/c!/w=10/abc.jpg", nil)
	imgReq := &ImageRequest{
		HTTPRequest: req,
	}

	originSlug, err := OriginSlugDetectFunc_Path(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin slug from URL path: %s. (req=%+v)", err.Error(), req)
	}

	if originSlug != originSlugExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originSlugExpected, originSlug, req)
	}
}

func TestOriginSlugDetect_Query(t *testing.T) {
	opts := ServerOptions{}
	var originSlugExpected OriginSlug = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/c!/w=10/abc.jpg?origin=klj8a", nil)
	imgReq := &ImageRequest{
		HTTPRequest: req,
	}

	originSlug, err := OriginSlugDetectFunc_Query(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin slug from query string: %s. (req=%+v)", err.Error(), req)
	}

	if originSlug != originSlugExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originSlugExpected, originSlug, req)
	}
}

func TestOriginSlugDetect_Header(t *testing.T) {
	opts := ServerOptions{}
	var originSlugExpected OriginSlug = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/c!/w=10/abc.jpg", nil)
	req.Header.Set(OriginSlugHTTPHeaderName, "klj8a")
	imgReq := &ImageRequest{
		HTTPRequest: req,
	}

	originSlug, err := OriginSlugDetectFunc_Header(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin slug from HTTP header: %s. (req=%+v)", err.Error(), req)
	}

	if originSlug != originSlugExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originSlugExpected, originSlug, req)
	}
}

func TestOriginSlugDetect_URLSignature(t *testing.T) {
	opts := ServerOptions{}
	var originSlugExpected OriginSlug = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/c!/w=10/abc.jpg?sig=klj8a-1.yiKX5u2kw6wp9zDgbrt2iOIi8IsoRIpw8fVgVc0yrNg", nil)
	sigInfo, err := parseURLSignature(req)
	imgReq := &ImageRequest{
		HTTPRequest:      req,
		URLSignatureInfo: sigInfo,
	}

	originSlug, err := OriginSlugDetectFunc_URLSignature(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin slug from URL signature: %s. (req=%+v)", err.Error(), req)
	}

	if originSlug != originSlugExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originSlugExpected, originSlug, req)
	}
}

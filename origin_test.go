package main

import (
	"net/http"
	"testing"
)

func TestOriginIdDetect_Host(t *testing.T) {
	opts := ServerOptions{
		OriginIdDetectHostPattern: `([a-z0-9]+)\.example\.test`,
	}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://klj8a.example.test/", nil)
	req.Host = req.URL.Host
	imgReq := &ImageRequest{
		HTTPRequest: req,
	}

	originId, err := OriginIdDetectFunc_Host(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin id from host name: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

func TestOriginIdDetect_Path(t *testing.T) {
	opts := ServerOptions{
		OriginIdDetectPathPattern: `^/([a-z0-9]+)/c!/`,
	}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/klj8a/c!/w=10/abc.jpg", nil)
	imgReq := &ImageRequest{
		HTTPRequest: req,
	}

	originId, err := OriginIdDetectFunc_Path(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin id from URL path: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

func TestOriginIdDetect_Query(t *testing.T) {
	opts := ServerOptions{}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/c!/w=10/abc.jpg?oid=klj8a", nil)
	imgReq := &ImageRequest{
		HTTPRequest: req,
	}

	originId, err := OriginIdDetectFunc_Query(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin id from query string: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

func TestOriginIdDetect_Header(t *testing.T) {
	opts := ServerOptions{}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/c!/w=10/abc.jpg", nil)
	req.Header.Set(OriginIdHTTPHeaderName, "klj8a")
	imgReq := &ImageRequest{
		HTTPRequest: req,
	}

	originId, err := OriginIdDetectFunc_Header(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin id from HTTP header: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

func TestOriginIdDetect_URLSignature(t *testing.T) {
	opts := ServerOptions{}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/c!/w=10/abc.jpg?sig=klj8a-1.yiKX5u2kw6wp9zDgbrt2iOIi8IsoRIpw8fVgVc0yrNg=", nil)
	sigInfo, err := parseURLSignature(req)
	imgReq := &ImageRequest{
		HTTPRequest:      req,
		URLSignatureInfo: sigInfo,
	}

	originId, err := OriginIdDetectFunc_URLSignature(opts, imgReq)
	if err != nil {
		t.Errorf("Failed to detect origin id from URL signature: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

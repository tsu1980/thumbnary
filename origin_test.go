package main

import (
	"net/http"
	"testing"
)

func TestOriginIdDetect_Host(t *testing.T) {
	sctx := &ServerContext{
		Options: ServerOptions{
			OriginIdDetectHostPattern: `([a-z0-9]+)\.example\.test`,
		},
	}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://klj8a.example.test/", nil)
	req.Host = req.URL.Host

	originId, err := OriginIdDetectFunc_Host(sctx, req)
	if err != nil {
		t.Errorf("Failed to detect origin id from host name: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

func TestOriginIdDetect_Path(t *testing.T) {
	sctx := &ServerContext{
		Options: ServerOptions{
			OriginIdDetectPathPattern: `^/([a-z0-9]+)/c!/`,
		},
	}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/klj8a/c!/w=10/abc.jpg", nil)

	originId, err := OriginIdDetectFunc_Path(sctx, req)
	if err != nil {
		t.Errorf("Failed to detect origin id from URL path: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

func TestOriginIdDetect_Query(t *testing.T) {
	sctx := &ServerContext{}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/c!/w=10/abc.jpg?oid=klj8a", nil)

	originId, err := OriginIdDetectFunc_Query(sctx, req)
	if err != nil {
		t.Errorf("Failed to detect origin id from query string: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

func TestOriginIdDetect_Header(t *testing.T) {
	sctx := &ServerContext{}
	var originIdExpected OriginId = "klj8a"
	req, _ := http.NewRequest("GET", "http://example.test/c!/w=10/abc.jpg", nil)
	req.Header.Set(OriginIdHTTPHeaderName, "klj8a")

	originId, err := OriginIdDetectFunc_Header(sctx, req)
	if err != nil {
		t.Errorf("Failed to detect origin id from HTTP header: %s. (req=%+v)", err.Error(), req)
	}

	if originId != originIdExpected {
		t.Errorf("Expected to '%s', but actual '%s'. (req=%+v)", originIdExpected, originId, req)
	}
}

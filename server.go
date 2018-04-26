package main

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

type ServerOptions struct {
	Port                      int
	Burst                     int
	Concurrency               int
	HTTPCacheTTL              int
	HTTPReadTimeout           int
	HTTPWriteTimeout          int
	MaxAllowedSize            int
	CORS                      bool
	AuthForwarding            bool
	EnablePlaceholder         bool
	EnableURLSignature        bool
	EnableOrigin              bool
	OriginIdDetectMethods     []OriginIdDetectMethod
	OriginIdDetectHostPattern string
	OriginIdDetectPathPattern string
	RedisURL                  string
	RedisChannelPrefix        string
	DBDriverName              string
	DBDataSourceName          string
	URLSignatureKey           string
	URLSignatureSalt          string
	Address                   string
	APIKey                    string
	CertFile                  string
	KeyFile                   string
	Authorization             string
	Placeholder               string
	PlaceholderImage          []byte
	OriginRepos               OriginRepository
}

func Server(o ServerOptions) error {
	addr := o.Address + ":" + strconv.Itoa(o.Port)
	handler := NewLog(NewHTTPHandler(o), os.Stdout)

	server := &http.Server{
		Addr:           addr,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    time.Duration(o.HTTPReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(o.HTTPWriteTimeout) * time.Second,
	}

	return listenAndServe(server, o)
}

func listenAndServe(s *http.Server, o ServerOptions) error {
	if o.CertFile != "" && o.KeyFile != "" {
		return s.ListenAndServeTLS(o.CertFile, o.KeyFile)
	}
	return s.ListenAndServe()
}

type MyHttpHandler struct {
	Options ServerOptions
}

func NewHTTPHandler(o ServerOptions) *MyHttpHandler {
	return &MyHttpHandler{Options: o}
}

func (h *MyHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		Middleware(indexController, h.Options).ServeHTTP(w, r)
	} else if r.URL.Path == "/health" {
		Middleware(healthController, h.Options).ServeHTTP(w, r)
	} else {
		ImageMiddleware(h.Options).ServeHTTP(w, r)
	}
}

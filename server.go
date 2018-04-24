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
}

type ServerContext struct {
	Options     ServerOptions
	OriginId    OriginId
	OriginRepos OriginRepository
}

func NewServerContext(o ServerOptions) *ServerContext {
	return &ServerContext{Options: o}
}

func Server(sctx *ServerContext) error {
	addr := sctx.Options.Address + ":" + strconv.Itoa(sctx.Options.Port)
	handler := NewLog(NewHTTPHandler(sctx), os.Stdout)

	server := &http.Server{
		Addr:           addr,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    time.Duration(sctx.Options.HTTPReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(sctx.Options.HTTPWriteTimeout) * time.Second,
	}

	return listenAndServe(server, sctx.Options)
}

func listenAndServe(s *http.Server, o ServerOptions) error {
	if o.CertFile != "" && o.KeyFile != "" {
		return s.ListenAndServeTLS(o.CertFile, o.KeyFile)
	}
	return s.ListenAndServe()
}

type MyHttpHandler struct {
	ServerContext *ServerContext
}

func NewHTTPHandler(sctx *ServerContext) *MyHttpHandler {
	return &MyHttpHandler{ServerContext: sctx}
}

func (h *MyHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		Middleware(indexController, h.ServerContext.Options).ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/health" {
		Middleware(healthController, h.ServerContext.Options).ServeHTTP(w, r)
		return
	}

	image := ImageMiddleware(h.ServerContext)
	image(ConvertImage).ServeHTTP(w, r)
}

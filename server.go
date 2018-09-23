package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type ServerOptions struct {
	Port                        int
	Burst                       int
	Concurrency                 int
	HTTPCacheTTL                int
	HTTPReadTimeout             int
	HTTPWriteTimeout            int
	MaxAllowedSize              int
	CORS                        bool
	AuthForwarding              bool
	EnablePlaceholder           bool
	EnableURLSignature          bool
	OriginSlugDetectMethods     []OriginSlugDetectMethod
	OriginSlugDetectHostPattern string
	OriginSlugDetectPathPattern string
	RedisURL                    string
	RedisChannelPrefix          string
	DBDriverName                string
	DBDataSourceName            string
	OriginTableName             string
	URLSignatureKey             string
	URLSignatureSalt            string
	Address                     string
	APIKey                      string
	CertFile                    string
	KeyFile                     string
	Authorization               string
	Placeholder                 string
	PlaceholderImage            []byte
	OriginRepos                 OriginRepository
}

func Server(o ServerOptions) {
	addr := o.Address + ":" + strconv.Itoa(o.Port)
	handler := NewLog(NewHTTPHandler(o), os.Stdout)

	server := &http.Server{
		Addr:           addr,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    time.Duration(o.HTTPReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(o.HTTPWriteTimeout) * time.Second,
	}

	listenAndServe(server, o)
}

func listenAndServe(s *http.Server, o ServerOptions) {
	go func() {
		var err error
		if o.CertFile != "" && o.KeyFile != "" {
			err = s.ListenAndServeTLS(o.CertFile, o.KeyFile)
		} else {
			err = s.ListenAndServe()
		}

		if err != nil {
			exitWithError("cannot start the server: %s", err)
		}
	}()

	// Wait for SIGINT/SIGTERM signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Printf("Shutdown start..")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Failed to shutdown %v", err)
	}
	log.Printf("Server gracefully stopped.")
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

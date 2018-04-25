package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	//"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/h2non/bimg.v1"
	"gopkg.in/h2non/filetype.v0"
)

type ImageRequest struct {
	Options ImageOptions
	Path    string
}

func indexController(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorReply(r, w, ErrNotFound, ServerOptions{})
		return
	}

	body, _ := json.Marshal(CurrentVersions)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func healthController(w http.ResponseWriter, r *http.Request) {
	health := GetHealthStats()
	body, _ := json.Marshal(health)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func imageController(sctx *ServerContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		imageSource := imageSourceMap[sctx.Origin.SourceType]
		if imageSource == nil {
			ErrorReply(req, w, ErrMissingImageSource, sctx.Options)
			return
		}

		r := regexp.MustCompile("/c!/([^/]+)/(.+)")
		values := r.FindStringSubmatch(req.URL.EscapedPath())
		if values == nil {
			err := fmt.Errorf("Bad URL format: %s", req.URL.EscapedPath())
			ErrorReply(req, w, NewError(err.Error(), BadRequest), sctx.Options)
			return
		}

		imgReq := &ImageRequest{}
		imgReq.Options = readParams(values[1])
		//log.Printf("readParams: %#v\n", opts)
		imgReq.Path = values[2]

		buf, err := imageSource.GetImage(req, sctx.Origin, imgReq.Path)
		if err != nil {
			ErrorReply(req, w, NewError(err.Error(), BadRequest), sctx.Options)
			return
		}

		if len(buf) == 0 {
			ErrorReply(req, w, ErrEmptyBody, sctx.Options)
			return
		}

		imageHandler(w, req, buf, imgReq, sctx.Options)
	}
}

func determineAcceptMimeType(accept string) string {
	for _, v := range strings.Split(accept, ",") {
		mediatype, _, _ := mime.ParseMediaType(v)
		if mediatype == "image/webp" {
			return "webp"
		} else if mediatype == "image/png" {
			return "png"
		} else if mediatype == "image/jpeg" {
			return "jpeg"
		}
	}
	// default
	return ""
}

func imageHandler(w http.ResponseWriter, r *http.Request, buf []byte, imgReq *ImageRequest, o ServerOptions) {
	// Infer the body MIME type via mimesniff algorithm
	mimeType := http.DetectContentType(buf)

	// If cannot infer the type, infer it via magic numbers
	if mimeType == "application/octet-stream" {
		kind, err := filetype.Get(buf)
		if err == nil && kind.MIME.Value != "" {
			mimeType = kind.MIME.Value
		}
	}

	// Infer text/plain responses as potential SVG image
	if strings.Contains(mimeType, "text/plain") && len(buf) > 8 {
		if bimg.IsSVGImage(buf) {
			mimeType = "image/svg+xml"
		}
	}

	// Finally check if image MIME type is supported
	if IsImageMimeTypeSupported(mimeType) == false {
		ErrorReply(r, w, ErrUnsupportedMedia, o)
		return
	}

	opts := imgReq.Options
	vary := ""
	if opts.Type == "auto" {
		opts.Type = determineAcceptMimeType(r.Header.Get("Accept"))
		vary = "Accept" // Ensure caches behave correctly for negotiated content
	} else if opts.Type != "" && ImageType(opts.Type) == 0 {
		ErrorReply(r, w, ErrOutputFormat, o)
		return
	}

	// Fetch overlay image if necessary
	if opts.NewOverlayURL != "" {
		var overlaySource = GetHttpSource()
		urlUnescaped, err := url.PathUnescape(opts.NewOverlayURL)
		if err != nil {
			ErrorReply(r, w, NewError(err.Error(), BadRequest), o)
			return
		}
		url, err := url.Parse(urlUnescaped)
		if err != nil {
			ErrorReply(r, w, NewError(err.Error(), BadRequest), o)
			return
		}

		//log.Printf("fetchImage overlay image: %#v", url.String())
		overlayBuf, err := overlaySource.fetchImage(url, r)
		if err != nil {
			ErrorReply(r, w, NewError(err.Error(), BadRequest), o)
			return
		}
		opts.NewOverlayBuf = overlayBuf
	}

	image, err := ConvertImage(buf, opts)
	if err != nil {
		ErrorReply(r, w, NewError("Error while processing the image: "+err.Error(), BadRequest), o)
		return
	}

	// Expose Content-Length response header
	w.Header().Set("Content-Length", strconv.Itoa(len(image.Body)))
	w.Header().Set("Content-Type", image.Mime)
	if vary != "" {
		w.Header().Set("Vary", vary)
	}
	w.Write(image.Body)
}

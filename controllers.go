package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/h2non/bimg.v1"
	"gopkg.in/h2non/filetype.v0"
)

type ImageRequest struct {
	HTTPRequest      *http.Request
	OriginSlug       OriginSlug
	Origin           *Origin
	Options          ImageOptions
	RelativeFilePath string
	URLSignatureInfo URLSignatureInfo
}

type URLSignatureInfo struct {
	Version        int
	SignatureValue string
	OriginSlug     OriginSlug
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

func imageController(o ServerOptions) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		imgReq := &ImageRequest{HTTPRequest: req}

		var err error
		imgReq.URLSignatureInfo, err = parseURLSignature(req)
		if err != nil {
			ErrorReply(req, w, NewError(err.Error(), BadRequest), o)
			return
		}

		_, err = FindOrigin(imgReq, o)
		if err != nil {
			ErrorReply(req, w, NewError(err.Error(), BadRequest), o)
			return
		}

		if imgReq.Origin.URLSignatureEnabled {
			err2 := validateURLSignature(imgReq)
			if err2 != nil {
				ErrorReply(req, w, *err2, o)
				return
			}
		}

		imageHandler(w, req, imgReq, o)
	}
}

func imageHandler(w http.ResponseWriter, req *http.Request, imgReq *ImageRequest, o ServerOptions) {
	imageSource := imageSourceMap[imgReq.Origin.SourceType]
	if imageSource == nil {
		ErrorReply(req, w, ErrMissingImageSource, o)
		return
	}

	r := regexp.MustCompile("/c!/([^/]+)/(.+)")
	values := r.FindStringSubmatch(req.URL.EscapedPath())
	if values == nil {
		err := fmt.Errorf("Bad URL format: %s", req.URL.EscapedPath())
		ErrorReply(req, w, NewError(err.Error(), BadRequest), o)
		return
	}

	imgReq.Options = readParams(values[1])
	//log.Printf("readParams: %#v\n", imgReq.Options)
	imgReq.RelativeFilePath = values[2]

	buf, err := imageSource.GetImage(req, imgReq.Origin, imgReq.RelativeFilePath)
	if err != nil {
		ErrorReply(req, w, NewError(err.Error(), BadRequest), o)
		return
	}

	if len(buf) == 0 {
		ErrorReply(req, w, ErrEmptyBody, o)
		return
	}

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
		ErrorReply(req, w, ErrUnsupportedMedia, o)
		return
	}

	opts := imgReq.Options
	vary := ""
	if opts.OutputFormat == "auto" {
		opts.OutputFormat = determineAcceptMimeType(req.Header.Get("Accept"))
		vary = "Accept" // Ensure caches behave correctly for negotiated content
	} else if opts.OutputFormat != "" && ImageType(opts.OutputFormat) == 0 {
		ErrorReply(req, w, ErrOutputFormat, o)
		return
	}

	// Fetch overlay image if necessary
	if opts.OverlayURL != "" {
		var overlaySource = GetHttpSource()
		urlUnescaped, err := url.PathUnescape(opts.OverlayURL)
		if err != nil {
			ErrorReply(req, w, NewError(err.Error(), BadRequest), o)
			return
		}
		url, err := url.Parse(urlUnescaped)
		if err != nil {
			ErrorReply(req, w, NewError(err.Error(), BadRequest), o)
			return
		}

		//log.Printf("fetchImage overlay image: %#v", url.String())
		overlayBuf, err := overlaySource.fetchImage(url, req)
		if err != nil {
			ErrorReply(req, w, NewError(err.Error(), BadRequest), o)
			return
		}
		opts.OverlayBuf = overlayBuf
	}

	imageFunc := ConvertImage
	if req.Method == "HEAD" {
		imageFunc = InfoImage
	}
	image, err := imageFunc(buf, opts)
	if err != nil {
		ErrorReply(req, w, NewError("Error while processing the image: "+err.Error(), BadRequest), o)
		return
	}

	// Expose Content-Length response header
	w.Header().Set("Content-Length", strconv.Itoa(len(image.Body)))
	w.Header().Set("Content-Type", image.Mime)
	if req.Method == "HEAD" {
		w.Header()["X-THUMBNARY-METADATA"] = []string{string(image.Body)}
	} else {
		if vary != "" {
			w.Header().Set("Vary", vary)
		}
		w.Write(image.Body)
	}
}

// version := "1"
// value := BASE64URL(HMAC-SHA-256(SigningKey, Path))
// originSlug := "ks8vm" + "-"	// Optional
// signature := originSlug + version + "." + value
//
// CreateURLSignatureString(1, "/c!/w=300/testdata/large.jpg", "secrettext", "")
func CreateURLSignatureString(version int, path string, key string, originSlug OriginSlug) string {
	var b strings.Builder

	if originSlug != "" {
		b.WriteString(string(originSlug) + "-")
	}

	b.WriteString(strconv.Itoa(version) + ".")

	sigVal := CalcURLSignatureValue(path, key)
	b.WriteString(sigVal)

	return b.String()
}

func CalcURLSignatureValue(path string, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(path))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

func parseURLSignature(req *http.Request) (URLSignatureInfo, error) {
	sigVal := req.URL.Query().Get("sig")
	if sigVal == "" {
		return URLSignatureInfo{}, nil
	}

	// parse URL signature string
	r := regexp.MustCompile(`([a-z0-9]+-)?(\d+)\.([A-Za-z0-9-_=]+)`)
	m := r.FindStringSubmatch(sigVal)
	mlen := len(m)
	if m == nil || (mlen != 4 && mlen != 3) {
		return URLSignatureInfo{}, fmt.Errorf("Bad URL signature format: %s", sigVal)
	}

	sigInfo := URLSignatureInfo{}

	if mlen == 4 {
		d, _ := strconv.Atoi(m[2])
		sigInfo.Version = d
		sigInfo.SignatureValue = m[3]
		sigInfo.OriginSlug = OriginSlug(strings.TrimSuffix(m[1], "-"))
	} else {
		d, _ := strconv.Atoi(m[1])
		sigInfo.Version = d
		sigInfo.SignatureValue = m[2]
	}

	return sigInfo, nil
}

func validateURLSignature(imgReq *ImageRequest) *Error {
	if imgReq.URLSignatureInfo.Version < 1 || imgReq.URLSignatureInfo.SignatureValue == "" {
		return &ErrInvalidURLSignature
	}

	// Compute expected URL signature
	var sigKey string
	if imgReq.URLSignatureInfo.Version == imgReq.Origin.URLSignatureKey_Version {
		sigKey = imgReq.Origin.URLSignatureKey
	} else if imgReq.URLSignatureInfo.Version == (imgReq.Origin.URLSignatureKey_Version - 1) {
		sigKey = imgReq.Origin.URLSignatureKey_Previous
	} else {
		return &ErrURLSignatureExpired
	}
	sigValExpected := CalcURLSignatureValue(imgReq.HTTPRequest.URL.EscapedPath(), sigKey)

	if strings.Compare(imgReq.URLSignatureInfo.SignatureValue, sigValExpected) != 0 {
		return &ErrURLSignatureMismatch
	}

	return nil
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

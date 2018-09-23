package main

import (
	"fmt"
	"net"
	"regexp"
)

type OriginRepositoryType string
type OriginSlug string

type Origin struct {
	Slug                     OriginSlug
	SourceType               ImageSourceType
	Scheme                   string
	Host                     string
	PathPrefix               string
	URLSignatureEnabled      bool
	URLSignatureKey          string
	URLSignatureKey_Previous string
	URLSignatureKey_Version  int
}

type OriginRepository interface {
	Open() error
	Close()
	Get(originSlug OriginSlug) (*Origin, error)
}

func NewOriginRepository(ort OriginRepositoryType, o ServerOptions) (OriginRepository, error) {
	switch ort {
	case OriginRepositoryTypeMySQL:
		return &MySQLOriginRepository{Options: o}, nil
	default:
		return nil, fmt.Errorf("Unknown repository type: (type=%s)", ort)
	}
}

type OriginSlugDetectMethod string
type OriginSlugDetectFunc func(ServerOptions, *ImageRequest) (OriginSlug, error)

const OriginSlugDetectMethod_Host OriginSlugDetectMethod = "host"
const OriginSlugDetectMethod_Path OriginSlugDetectMethod = "path"
const OriginSlugDetectMethod_Query OriginSlugDetectMethod = "query"
const OriginSlugDetectMethod_Header OriginSlugDetectMethod = "header"
const OriginSlugDetectMethod_URLSignature OriginSlugDetectMethod = "urlsig"

const OriginSlugHTTPHeaderName string = "X-THUMBNARY-ORIGIN-SLUG"

var originSlugDetectMethodMap = make(map[OriginSlugDetectMethod]OriginSlugDetectFunc)

func OriginSlugDetectFunc_Host(o ServerOptions, imgReq *ImageRequest) (OriginSlug, error) {
	host, _, err := net.SplitHostPort(imgReq.HTTPRequest.Host)
	if err != nil {
		host = imgReq.HTTPRequest.Host
	}

	r := regexp.MustCompile(o.OriginSlugDetectHostPattern)
	group := r.FindStringSubmatch(host)
	if group == nil {
		return "", fmt.Errorf("Cannot extract origin slug: (host=%s)", host)
	}

	var originSlug = OriginSlug(group[1])
	if originSlug == "" {
		return "", fmt.Errorf("Origin slug is empty: (host=%s)", host)
	}

	return originSlug, nil
}

func OriginSlugDetectFunc_Path(o ServerOptions, imgReq *ImageRequest) (OriginSlug, error) {
	r := regexp.MustCompile(o.OriginSlugDetectPathPattern)
	group := r.FindStringSubmatch(imgReq.HTTPRequest.URL.Path)
	if group == nil {
		return "", fmt.Errorf("Cannot extract origin slug: (path=%s)", imgReq.HTTPRequest.URL.Path)
	}

	var originSlug = OriginSlug(group[1])
	if originSlug == "" {
		return "", fmt.Errorf("Origin slug is empty: (path=%s)", imgReq.HTTPRequest.URL.Path)
	}

	return originSlug, nil
}

func OriginSlugDetectFunc_Query(o ServerOptions, imgReq *ImageRequest) (OriginSlug, error) {
	originSlug := imgReq.HTTPRequest.URL.Query().Get("origin")
	if originSlug == "" {
		return "", fmt.Errorf("origin query string not specified: (URL=%s)", imgReq.HTTPRequest.URL.String())
	}
	return (OriginSlug)(originSlug), nil
}

func OriginSlugDetectFunc_Header(o ServerOptions, imgReq *ImageRequest) (OriginSlug, error) {
	originSlug := imgReq.HTTPRequest.Header.Get(OriginSlugHTTPHeaderName)
	if originSlug == "" {
		return "", fmt.Errorf("Origin slug in HTTP header is not specified: (Header=%+v)", imgReq.HTTPRequest.Header)
	}
	return (OriginSlug)(originSlug), nil
}

func OriginSlugDetectFunc_URLSignature(o ServerOptions, imgReq *ImageRequest) (OriginSlug, error) {
	if imgReq.URLSignatureInfo.OriginSlug == "" {
		return "", fmt.Errorf("Origin slug in URL signature is not specified: (URL=%s)", imgReq.HTTPRequest.URL.String())
	}
	return imgReq.URLSignatureInfo.OriginSlug, nil
}

func FindOrigin(imgReq *ImageRequest, o ServerOptions) (*Origin, error) {
	for _, method := range o.OriginSlugDetectMethods {
		methodFunc, _ := originSlugDetectMethodMap[method]

		originSlug, _ := methodFunc(o, imgReq)
		if originSlug != "" {
			//log.Printf("Origin slug is %s", (string)(originSlug))

			origin, err := o.OriginRepos.Get(originSlug)
			if err != nil {
				return nil, err
			}

			imgReq.OriginSlug = originSlug
			imgReq.Origin = origin
			return origin, nil
		}
	}

	return nil, fmt.Errorf("Cannot detect origin slug: (methods=%+v) (req=%+v)", o.OriginSlugDetectMethods, imgReq)
}

func init() {
	originSlugDetectMethodMap[OriginSlugDetectMethod_Host] = OriginSlugDetectFunc_Host
	originSlugDetectMethodMap[OriginSlugDetectMethod_Path] = OriginSlugDetectFunc_Path
	originSlugDetectMethodMap[OriginSlugDetectMethod_Query] = OriginSlugDetectFunc_Query
	originSlugDetectMethodMap[OriginSlugDetectMethod_Header] = OriginSlugDetectFunc_Header
	originSlugDetectMethodMap[OriginSlugDetectMethod_URLSignature] = OriginSlugDetectFunc_URLSignature
}

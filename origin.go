package main

import (
	"fmt"
	"net"
	"regexp"
)

type OriginRepositoryType string
type OriginId string

type Origin struct {
	ID                       OriginId
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
	Get(originId OriginId) (*Origin, error)
}

func NewOriginRepository(ort OriginRepositoryType, o ServerOptions) (OriginRepository, error) {
	switch ort {
	case OriginRepositoryTypeMySQL:
		return &MySQLOriginRepository{Options: o}, nil
	default:
		return nil, fmt.Errorf("Unknown repository type: (type=%s)", ort)
	}
}

type OriginIdDetectMethod string
type OriginIdDetectFunc func(ServerOptions, *ImageRequest) (OriginId, error)

const OriginIdDetectMethod_Host OriginIdDetectMethod = "host"
const OriginIdDetectMethod_Path OriginIdDetectMethod = "path"
const OriginIdDetectMethod_Query OriginIdDetectMethod = "query"
const OriginIdDetectMethod_Header OriginIdDetectMethod = "header"
const OriginIdDetectMethod_URLSignature OriginIdDetectMethod = "urlsig"

const OriginIdHTTPHeaderName string = "X-THUMBNARY-ORIGIN-ID"

var originIdDetectMethodMap = make(map[OriginIdDetectMethod]OriginIdDetectFunc)

func OriginIdDetectFunc_Host(o ServerOptions, imgReq *ImageRequest) (OriginId, error) {
	host, _, err := net.SplitHostPort(imgReq.HTTPRequest.Host)
	if err != nil {
		host = imgReq.HTTPRequest.Host
	}

	r := regexp.MustCompile(o.OriginIdDetectHostPattern)
	group := r.FindStringSubmatch(host)
	if group == nil {
		return "", fmt.Errorf("Cannot extract origin id: (host=%s)", host)
	}

	var originId = OriginId(group[1])
	if originId == "" {
		return "", fmt.Errorf("Origin id is empty: (host=%s)", host)
	}

	return originId, nil
}

func OriginIdDetectFunc_Path(o ServerOptions, imgReq *ImageRequest) (OriginId, error) {
	r := regexp.MustCompile(o.OriginIdDetectPathPattern)
	group := r.FindStringSubmatch(imgReq.HTTPRequest.URL.Path)
	if group == nil {
		return "", fmt.Errorf("Cannot extract origin id: (path=%s)", imgReq.HTTPRequest.URL.Path)
	}

	var originId = OriginId(group[1])
	if originId == "" {
		return "", fmt.Errorf("Origin id is empty: (path=%s)", imgReq.HTTPRequest.URL.Path)
	}

	return originId, nil
}

func OriginIdDetectFunc_Query(o ServerOptions, imgReq *ImageRequest) (OriginId, error) {
	originId := imgReq.HTTPRequest.URL.Query().Get("oid")
	if originId == "" {
		return "", fmt.Errorf("oid query string not specified: (URL=%s)", imgReq.HTTPRequest.URL.String())
	}
	return (OriginId)(originId), nil
}

func OriginIdDetectFunc_Header(o ServerOptions, imgReq *ImageRequest) (OriginId, error) {
	originId := imgReq.HTTPRequest.Header.Get(OriginIdHTTPHeaderName)
	if originId == "" {
		return "", fmt.Errorf("Origin id in HTTP header is not specified: (Header=%+v)", imgReq.HTTPRequest.Header)
	}
	return (OriginId)(originId), nil
}

func OriginIdDetectFunc_URLSignature(o ServerOptions, imgReq *ImageRequest) (OriginId, error) {
	if imgReq.URLSignatureInfo.OriginId == "" {
		return "", fmt.Errorf("Origin id in URL signature is not specified: (URL=%s)", imgReq.HTTPRequest.URL.String())
	}
	return imgReq.URLSignatureInfo.OriginId, nil
}

func FindOrigin(imgReq *ImageRequest, o ServerOptions) (*Origin, error) {
	for _, method := range o.OriginIdDetectMethods {
		methodFunc, _ := originIdDetectMethodMap[method]

		originId, _ := methodFunc(o, imgReq)
		if originId != "" {
			//log.Printf("Origin id is %s", (string)(originId))

			origin, err := o.OriginRepos.Get(originId)
			if err != nil {
				return nil, err
			}

			imgReq.OriginId = originId
			imgReq.Origin = origin
			return origin, nil
		}
	}

	return nil, fmt.Errorf("Cannot detect origin id: (methods=%+v) (req=%+v)", o.OriginIdDetectMethods, imgReq)
}

func init() {
	originIdDetectMethodMap[OriginIdDetectMethod_Host] = OriginIdDetectFunc_Host
	originIdDetectMethodMap[OriginIdDetectMethod_Path] = OriginIdDetectFunc_Path
	originIdDetectMethodMap[OriginIdDetectMethod_Query] = OriginIdDetectFunc_Query
	originIdDetectMethodMap[OriginIdDetectMethod_Header] = OriginIdDetectFunc_Header
	originIdDetectMethodMap[OriginIdDetectMethod_URLSignature] = OriginIdDetectFunc_URLSignature
}

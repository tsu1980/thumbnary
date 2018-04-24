package main

import (
	"fmt"
	"net"
	"net/http"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

type OriginRepositoryType string
type OriginId string

type Origin struct {
	ID                       OriginId
	SourceType               ImageSourceType
	Scheme                   string
	Host                     string
	PathPrefix               string
	URLSignatureKey          string
	URLSignatureKey_Previous string
	URLSignatureKey_Version  string
}

type OriginRepository interface {
	Open() error
	Close()
	Get(originId OriginId) (*Origin, error)
}

func NewOriginRepository(ort OriginRepositoryType, o ServerOptions) (OriginRepository, error) {
	switch ort {
	case "mysql":
		return &MySQLOriginRepository{Options: o}, nil
	default:
		return nil, fmt.Errorf("Unknown repository type: (type=%s)", ort)
	}
}

type OriginIdDetectMethod string
type OriginIdDetectFunc func(*ServerContext, *http.Request) (OriginId, error)

const OriginIdDetectMethod_Host OriginIdDetectMethod = "host"
const OriginIdDetectMethod_Path OriginIdDetectMethod = "path"
const OriginIdDetectMethod_Query OriginIdDetectMethod = "query"
const OriginIdDetectMethod_Header OriginIdDetectMethod = "header"
const OriginIdDetectMethod_URLSignature OriginIdDetectMethod = "urlsig"

const OriginIdHTTPHeaderName string = "X-THUMBNARY-ORIGIN-ID"

var originIdDetectMethodMap = make(map[OriginIdDetectMethod]OriginIdDetectFunc)

func OriginIdDetectFunc_Host(sctx *ServerContext, req *http.Request) (OriginId, error) {
	host, _, err := net.SplitHostPort(req.Host)
	if err != nil {
		host = req.Host
	}

	r := regexp.MustCompile(sctx.Options.OriginIdDetectHostPattern)
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

func OriginIdDetectFunc_Path(sctx *ServerContext, req *http.Request) (OriginId, error) {
	r := regexp.MustCompile(sctx.Options.OriginIdDetectPathPattern)
	group := r.FindStringSubmatch(req.URL.Path)
	if group == nil {
		return "", fmt.Errorf("Cannot extract origin id: (path=%s)", req.URL.Path)
	}

	var originId = OriginId(group[1])
	if originId == "" {
		return "", fmt.Errorf("Origin id is empty: (path=%s)", req.URL.Path)
	}

	return originId, nil
}

func OriginIdDetectFunc_Query(sctx *ServerContext, req *http.Request) (OriginId, error) {
	originId := req.URL.Query().Get("oid")
	if originId == "" {
		return "", fmt.Errorf("oid query string not specified: (URL=%s)", req.URL.String())
	}
	return (OriginId)(originId), nil
}

func OriginIdDetectFunc_Header(sctx *ServerContext, req *http.Request) (OriginId, error) {
	originId := req.Header.Get(OriginIdHTTPHeaderName)
	if originId == "" {
		return "", fmt.Errorf("oid query not specified: (Header=%+v)", req.Header)
	}
	return (OriginId)(originId), nil
}

func OriginIdDetectFunc_URLSignature(sctx *ServerContext, req *http.Request) (OriginId, error) {
	return "", fmt.Errorf("Not implemented yet")
}

func FindOrigin(sctx *ServerContext, req *http.Request) (*Origin, error) {
	for _, method := range sctx.Options.OriginIdDetectMethods {
		methodFunc, _ := originIdDetectMethodMap[method]

		originId, _ := methodFunc(sctx, req)
		if originId != "" {
			sctx.OriginId = originId
			//log.Printf("Origin id is %s", (string)(originId))

			origin, err := sctx.OriginRepos.Get(originId)
			if err != nil {
				return nil, err
			}
			return origin, nil
		}
	}

	return nil, fmt.Errorf("Cannot detect origin id: (req=%#v)", req)
}

func init() {
	originIdDetectMethodMap[OriginIdDetectMethod_Host] = OriginIdDetectFunc_Host
	originIdDetectMethodMap[OriginIdDetectMethod_Path] = OriginIdDetectFunc_Path
	originIdDetectMethodMap[OriginIdDetectMethod_Query] = OriginIdDetectFunc_Query
	originIdDetectMethodMap[OriginIdDetectMethod_Header] = OriginIdDetectFunc_Header
	originIdDetectMethodMap[OriginIdDetectMethod_URLSignature] = OriginIdDetectFunc_URLSignature
}

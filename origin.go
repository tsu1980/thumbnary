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

func FindOrigin(sctx *ServerContext, req *http.Request) (*Origin, error) {
	host, _, _ := net.SplitHostPort(req.Host)

	r := regexp.MustCompile(sctx.Options.OriginHostPattern)
	group := r.FindStringSubmatch(host)
	if group == nil {
		return nil, fmt.Errorf("Cannot extract origin id: (host=%s)", host)
	}

	var originId = OriginId(group[1])
	if originId == "" {
		return nil, fmt.Errorf("Origin id is empty: (host=%s)", host)
	}

	origin, err := sctx.OriginRepos.Get(originId)
	if err != nil {
		return nil, err
	}
	return origin, nil
}

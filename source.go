package main

import (
	"net/http"
)

type ImageSourceType string
type ImageSourceFactoryFunction func(*SourceConfig) ImageSource

type SourceConfig struct {
	AuthForwarding bool
	Authorization  string
	Type           ImageSourceType
	MaxAllowedSize int
}

var imageSourceMap = make(map[ImageSourceType]ImageSource)
var imageSourceFactoryMap = make(map[ImageSourceType]ImageSourceFactoryFunction)

type ImageSource interface {
	Matches(*http.Request) bool
	GetImage(*http.Request, *Origin) ([]byte, error)
}

func RegisterSource(sourceType ImageSourceType, factory ImageSourceFactoryFunction) {
	imageSourceFactoryMap[sourceType] = factory
}

func LoadSources(o ServerOptions) {
	for name, factory := range imageSourceFactoryMap {
		imageSourceMap[name] = factory(&SourceConfig{
			Type:           name,
			AuthForwarding: o.AuthForwarding,
			Authorization:  o.Authorization,
			MaxAllowedSize: o.MaxAllowedSize,
		})
	}
}

func MatchSource(req *http.Request) ImageSource {
	for _, source := range imageSourceMap {
		if source.Matches(req) {
			return source
		}
	}
	return nil
}

func GetHttpSource() *HttpImageSource {
	return imageSourceMap[ImageSourceTypeHttp].(*HttpImageSource)
}

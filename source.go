package main

import "net/http"

type ImageSourceType int

const ImageSourceTypeHttp ImageSourceType = 0

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
	GetImage(*http.Request, *Origin, string) ([]byte, error)
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

func GetHttpSource() *HttpImageSource {
	return imageSourceMap[ImageSourceTypeHttp].(*HttpImageSource)
}

package main

import "gopkg.in/h2non/bimg.v1"

type ResizeMode int

const (
	ResizeModeScale ResizeMode = 0
	ResizeModeCrop  ResizeMode = 1
	ResizeModeFit   ResizeMode = 2
	ResizeModePad   ResizeMode = 4
)

type Gravity9 int

const (
	Gravity9TopLeft      Gravity9 = 1
	Gravity9TopCenter    Gravity9 = 2
	Gravity9TopRight     Gravity9 = 3
	Gravity9MiddleLeft   Gravity9 = 4
	Gravity9MiddleCenter Gravity9 = 5
	Gravity9MiddleRight  Gravity9 = 6
	Gravity9BottomLeft   Gravity9 = 7
	Gravity9BottomCenter Gravity9 = 8
	Gravity9BottomRight  Gravity9 = 9
	Gravity9Smart        Gravity9 = 20
)

// ImageOptions represent all the supported image transformation params as first level members
type ImageOptions struct {
	Width         int
	Height        int
	AreaWidth     int
	AreaHeight    int
	Quality       int
	Compression   int
	Rotate        int
	Top           int
	Left          int
	Margin        int
	Factor        int
	DPI           int
	TextWidth     int
	Flip          bool
	Flop          bool
	Force         bool
	Embed         bool
	NoCrop        bool
	NoReplicate   bool
	NoRotation    bool
	NoProfile     bool
	StripMetadata bool
	Opacity       float32
	Sigma         float64
	MinAmpl       float64
	Text          string
	Font          string
	Type          string
	Color         []uint8
	Background    []uint8
	Extend        bimg.Extend
	Gravity       bimg.Gravity
	Colorspace    bimg.Interpretation

	// for new url format
	NewWidth      int
	NewHeight     int
	NewUpscale    bool
	NewResizeMode ResizeMode
	//	NewClip       []int
	//	NewClipRate   []float
	NewGravity    Gravity9
	NewBackground []uint8

	NewOverlayURL     string
	NewOverlayBuf     []byte
	NewOverlayX       int
	NewOverlayY       int
	NewOverlayGravity Gravity9
	NewOverlayOpacity float32

	NewMonochrome bool

	NewFileType string
	NewQuality  int
}

// BimgOptions creates a new bimg compatible options struct mapping the fields properly
func BimgOptions(o ImageOptions) bimg.Options {
	opts := bimg.Options{
		Width:          o.Width,
		Height:         o.Height,
		Flip:           o.Flip,
		Flop:           o.Flop,
		Quality:        o.Quality,
		Compression:    o.Compression,
		NoAutoRotate:   o.NoRotation,
		NoProfile:      o.NoProfile,
		Force:          o.Force,
		Gravity:        o.Gravity,
		Embed:          o.Embed,
		Extend:         o.Extend,
		Interpretation: o.Colorspace,
		StripMetadata:  o.StripMetadata,
		Type:           ImageType(o.Type),
		Rotate:         bimg.Angle(o.Rotate),
	}

	if len(o.Background) != 0 {
		opts.Background = bimg.Color{o.Background[0], o.Background[1], o.Background[2]}
	}

	if o.Sigma > 0 || o.MinAmpl > 0 {
		opts.GaussianBlur = bimg.GaussianBlur{
			Sigma:   o.Sigma,
			MinAmpl: o.MinAmpl,
		}
	}

	return opts
}

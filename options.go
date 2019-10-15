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
	NoConvert  bool
	Width      int
	Height     int
	Upscale    bool
	ResizeMode ResizeMode
	//	Clip       []int
	//	ClipRate   []float
	Gravity    Gravity9
	Background []uint8

	OverlayURL     string
	OverlayBuf     []byte
	OverlayX       int
	OverlayY       int
	OverlayGravity Gravity9
	OverlayOpacity float32

	Monochrome bool

	OutputFormat string
	Quality      int
}

// ImageOptionsNoConvert represent No conversion options
var ImageOptionsNoConvert = ImageOptions{
	NoConvert: true,
}

// BimgOptions creates a new bimg compatible options struct mapping the fields properly
func BimgOptions(o ImageOptions) bimg.Options {
	opts := bimg.Options{
		Width:          o.Width,
		Height:         o.Height,
		Flip:           false,
		Flop:           false,
		Quality:        o.Quality,
		Compression:    6,
		NoAutoRotate:   false,
		NoProfile:      false,
		Force:          false,
		Gravity:        bimg.GravityCentre,
		Embed:          false,
		Extend:         bimg.ExtendBlack,
		Interpretation: bimg.InterpretationSRGB,
		StripMetadata:  true,
		Type:           ImageType(o.OutputFormat),
		Rotate:         bimg.Angle(0),
	}

	if len(o.Background) != 0 {
		opts.Background = bimg.Color{o.Background[0], o.Background[1], o.Background[2]}
		opts.Extend = bimg.ExtendBackground
	}
	if o.Upscale {
		opts.Enlarge = true
	}

	var m = map[Gravity9]bimg.Gravity{
		Gravity9BottomCenter: bimg.GravitySouth,
		Gravity9TopCenter:    bimg.GravityNorth,
		Gravity9MiddleRight:  bimg.GravityEast,
		Gravity9MiddleLeft:   bimg.GravityWest,
		Gravity9MiddleCenter: bimg.GravityCentre,
		Gravity9Smart:        bimg.GravitySmart,
	}
	if g, ok := m[o.Gravity]; ok {
		opts.Gravity = g
	}

	if o.Monochrome {
		opts.Interpretation = bimg.InterpretationBW
	}

	if o.OverlayURL != "" {
		opts.WatermarkImage.Left = o.OverlayX
		opts.WatermarkImage.Top = o.OverlayY
		opts.WatermarkImage.Buf = o.OverlayBuf
		opts.WatermarkImage.Opacity = o.OverlayOpacity
	}

	return opts
}

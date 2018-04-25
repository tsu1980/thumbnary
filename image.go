package main

import (
	"encoding/json"
	"errors"

	"gopkg.in/h2non/bimg.v1"
)

// Image stores an image binary buffer and its MIME type
type Image struct {
	Body []byte
	Mime string
}

// ImageInfo represents an image details and additional metadata
type ImageInfo struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Type        string `json:"type"`
	Space       string `json:"space"`
	Alpha       bool   `json:"hasAlpha"`
	Profile     bool   `json:"hasProfile"`
	Channels    int    `json:"channels"`
	Orientation int    `json:"orientation"`
}

func Info(buf []byte, o ImageOptions) (Image, error) {
	// We're not handling an image here, but we reused the struct.
	// An interface will be definitively better here.
	image := Image{Mime: "application/json"}

	meta, err := bimg.Metadata(buf)
	if err != nil {
		return image, NewError("Cannot retrieve image metadata: %s"+err.Error(), BadRequest)
	}

	info := ImageInfo{
		Width:       meta.Size.Width,
		Height:      meta.Size.Height,
		Type:        meta.Type,
		Space:       meta.Space,
		Alpha:       meta.Alpha,
		Profile:     meta.Profile,
		Channels:    meta.Channels,
		Orientation: meta.Orientation,
	}

	body, _ := json.Marshal(info)
	image.Body = body

	return image, nil
}

func ConvertImage(buf []byte, o ImageOptions) (Image, error) {
	opts := BimgOptions(o)
	switch o.NewResizeMode {
	case ResizeModeCrop:
		if o.NewWidth == 0 && o.NewHeight == 0 {
			return Image{}, NewError("Missing required param: height or width", BadRequest)
		}
		opts.Crop = true
	case ResizeModeFit:
		if o.NewWidth == 0 || o.NewHeight == 0 {
			return Image{}, NewError("Missing required params: height, width", BadRequest)
		}
		dims, err := bimg.Size(buf)
		if err != nil {
			return Image{}, err
		}

		// if input ratio > output ratio
		// (calculation multiplied through by denominators to avoid float division)
		if dims.Width*o.NewHeight > o.NewWidth*dims.Height {
			// constrained by width
			if dims.Width != 0 {
				o.NewHeight = o.NewWidth * dims.Height / dims.Width
			}
		} else {
			// constrained by height
			if dims.Height != 0 {
				o.NewWidth = o.NewHeight * dims.Width / dims.Height
			}
		}

		//opts.Embed = true
	case ResizeModePad:
		if o.NewWidth == 0 && o.NewHeight == 0 {
			return Image{}, NewError("Missing required param: height or width", BadRequest)
		}
		opts.Embed = true
	case ResizeModeScale:
		if o.NewWidth == 0 && o.NewHeight == 0 {
			return Image{}, NewError("Missing required param: height or width", BadRequest)
		}
	}

	opts.Width = o.NewWidth
	opts.Height = o.NewHeight
	opts.Type = ImageType(o.Type)
	if len(o.NewBackground) != 0 {
		opts.Background = bimg.Color{o.NewBackground[0], o.NewBackground[1], o.NewBackground[2]}
		opts.Extend = bimg.ExtendBackground
	}
	if o.NewUpscale {
		opts.Enlarge = true
	}

	var m = map[Gravity9]bimg.Gravity{
		Gravity9BottomCenter: bimg.GravitySouth,
		Gravity9TopCenter:    bimg.GravityNorth,
		Gravity9MiddleRight:  bimg.GravityEast,
		Gravity9BottomLeft:   bimg.GravityWest,
		Gravity9MiddleCenter: bimg.GravityCentre,
		Gravity9Smart:        bimg.GravitySmart,
	}
	if g, ok := m[o.NewGravity]; ok {
		opts.Gravity = g
	}

	opts.StripMetadata = true

	if o.NewOverlayURL != "" {
		opts.WatermarkImage.Left = o.NewOverlayX
		opts.WatermarkImage.Top = o.NewOverlayY
		opts.WatermarkImage.Buf = o.NewOverlayBuf
		opts.WatermarkImage.Opacity = o.NewOverlayOpacity
	}

	return Process(buf, opts)
}

func Process(buf []byte, opts bimg.Options) (out Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch value := r.(type) {
			case error:
				err = value
			case string:
				err = errors.New(value)
			default:
				err = errors.New("libvips internal error")
			}
			out = Image{}
		}
	}()

	buf, err = bimg.Resize(buf, opts)
	if err != nil {
		return Image{}, err
	}

	mime := GetImageMimeType(bimg.DetermineImageType(buf))
	return Image{Body: buf, Mime: mime}, nil
}

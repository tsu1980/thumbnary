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
	Version int             `json:"version"`
	Source  ImageInfoSource `json:"source"`
}

type ImageInfoSource struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Type        string `json:"type"`
	Space       string `json:"space"`
	Alpha       bool   `json:"hasAlpha"`
	Profile     bool   `json:"hasProfile"`
	Channels    int    `json:"channels"`
	Orientation int    `json:"orientation"`
}

func InfoImage(buf []byte, o ImageOptions) (Image, error) {
	// We're not handling an image here, but we reused the struct.
	// An interface will be definitively better here.
	image := Image{Mime: "application/json"}

	meta, err := bimg.Metadata(buf)
	if err != nil {
		return image, NewError("Cannot retrieve image metadata: %s"+err.Error(), BadRequest)
	}

	info := ImageInfo{
		Version: 1,
		Source: ImageInfoSource{
			Width:       meta.Size.Width,
			Height:      meta.Size.Height,
			Type:        meta.Type,
			Space:       meta.Space,
			Alpha:       meta.Alpha,
			Profile:     meta.Profile,
			Channels:    meta.Channels,
			Orientation: meta.Orientation,
		},
	}

	body, _ := json.Marshal(info)
	image.Body = body

	return image, nil
}

func ConvertImage(buf []byte, o ImageOptions) (Image, error) {
	if o.NoConvert == true {
		mime := GetImageMimeType(bimg.DetermineImageType(buf))
		return Image{Body: buf, Mime: mime}, nil
	}

	opts := BimgOptions(o)

	switch o.ResizeMode {
	case ResizeModeCrop:
		if o.Width == 0 && o.Height == 0 {
			return Image{}, NewError("Missing required param: height or width", BadRequest)
		}
		opts.Crop = true
	case ResizeModeFit:
		if o.Width == 0 || o.Height == 0 {
			return Image{}, NewError("Missing required params: height, width", BadRequest)
		}
		dims, err := bimg.Size(buf)
		if err != nil {
			return Image{}, err
		}

		// if input ratio > output ratio
		// (calculation multiplied through by denominators to avoid float division)
		if dims.Width*o.Height > o.Width*dims.Height {
			// constrained by width
			if dims.Width != 0 {
				opts.Height = o.Width * dims.Height / dims.Width
			}
		} else {
			// constrained by height
			if dims.Height != 0 {
				opts.Width = o.Height * dims.Width / dims.Height
			}
		}

		//opts.Embed = true
	case ResizeModePad:
		if o.Width == 0 && o.Height == 0 {
			return Image{}, NewError("Missing required param: height or width", BadRequest)
		}
		opts.Embed = true
	case ResizeModeScale:
		if o.Width == 0 && o.Height == 0 {
			return Image{}, NewError("Missing required param: height or width", BadRequest)
		}
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

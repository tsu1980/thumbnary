package main

import (
	"io/ioutil"
	"testing"
)

func TestImageResize(t *testing.T) {
	opts := ImageOptions{
		NewWidth:      300,
		NewHeight:     300,
		NewResizeMode: ResizeModeCrop,
		Type:          "auto",
	}
	buf, _ := ioutil.ReadAll(readFile("thumbnary.jpg"))

	img, err := ConvertImage(buf, opts)
	if err != nil {
		t.Errorf("Cannot process image: %s", err)
	}
	if img.Mime != "image/jpeg" {
		t.Error("Invalid image MIME type")
	}
	if assertSize(img.Body, opts.NewWidth, opts.NewHeight) != nil {
		t.Errorf("Invalid image size, expected: %dx%d", opts.NewWidth, opts.NewHeight)
	}
}

func TestImageFit(t *testing.T) {
	opts := ImageOptions{
		NewWidth:      300,
		NewHeight:     300,
		NewResizeMode: ResizeModeFit,
	}
	buf, _ := ioutil.ReadAll(readFile("thumbnary.jpg"))

	img, err := ConvertImage(buf, opts)
	if err != nil {
		t.Errorf("Cannot process image: %s", err)
	}
	if img.Mime != "image/jpeg" {
		t.Error("Invalid image MIME type")
	}
	// 550x740 -> 222x300
	if assertSize(img.Body, 222, 300) != nil {
		t.Errorf("Invalid image size, expected: %dx%d", opts.NewWidth, opts.NewHeight)
	}
}

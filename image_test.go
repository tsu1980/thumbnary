package main

import (
	"io/ioutil"
	"testing"
)

func TestImageResize(t *testing.T) {
	opts := ImageOptions{
		Width:        300,
		Height:       300,
		ResizeMode:   ResizeModeCrop,
		OutputFormat: "auto",
	}
	buf, _ := ioutil.ReadAll(readFile("thumbnary.jpg"))

	img, err := ConvertImage(buf, opts)
	if err != nil {
		t.Errorf("Cannot process image: %s", err)
	}
	if img.Mime != "image/jpeg" {
		t.Error("Invalid image MIME type")
	}
	if assertSize(img.Body, opts.Width, opts.Height) != nil {
		t.Errorf("Invalid image size, expected: %dx%d", opts.Width, opts.Height)
	}
}

func TestImageFit(t *testing.T) {
	opts := ImageOptions{
		Width:      300,
		Height:     300,
		ResizeMode: ResizeModeFit,
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
	if err = assertSize(img.Body, 222, 300); err != nil {
		t.Errorf(err.Error())
	}
}

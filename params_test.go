package main

import (
	"testing"

	"gopkg.in/h2non/bimg.v1"
)

const fixture = "fixtures/large.jpg"

func TestReadParams(t *testing.T) {
	str := "w=100,h=80,lo=0.2,b=ff0a14"
	params := readParams(str)

	assert := params.Width == 100 &&
		params.Height == 80 &&
		params.OverlayOpacity == 0.2 &&
		params.Background[0] == 255 &&
		params.Background[1] == 10 &&
		params.Background[2] == 20

	if assert == false {
		t.Error("Invalid params")
	}
}

func TestParseParam(t *testing.T) {
	intCases := []struct {
		value    string
		expected int
	}{
		{"1", 1},
		{"0100", 100},
		{"-100", 100},
		{"99.02", 99},
		{"99.9", 100},
	}

	for _, test := range intCases {
		val := parseParam(test.value, "int")
		if val != test.expected {
			t.Errorf("Invalid param: %s != %d", test.value, test.expected)
		}
	}

	floatCases := []struct {
		value    string
		expected float64
	}{
		{"1.1", 1.1},
		{"01.1", 1.1},
		{"-1.10", 1.10},
		{"99.999999", 99.999999},
	}

	for _, test := range floatCases {
		val := parseParam(test.value, "float")
		if val != test.expected {
			t.Errorf("Invalid param: %#v != %#v", val, test.expected)
		}
	}

	boolCases := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"1.1", false},
		{"-1", false},
		{"0", false},
		{"0.0", false},
		{"no", false},
		{"yes", false},
	}

	for _, test := range boolCases {
		val := parseParam(test.value, "bool")
		if val != test.expected {
			t.Errorf("Invalid param: %#v != %#v", val, test.expected)
		}
	}
}

func TestParseColor(t *testing.T) {
	cases := []struct {
		value    string
		expected []uint8
	}{
		{"200,100,20", []uint8{200, 100, 20}},
		{"0,280,200", []uint8{0, 255, 200}},
		{" -1, 256 , 50", []uint8{0, 255, 50}},
		{" a, 20 , &hel0", []uint8{0, 20, 0}},
		{"", []uint8{}},
	}

	for _, color := range cases {
		c := parseColor(color.value)
		l := len(color.expected)

		if len(c) != l {
			t.Errorf("Invalid color length: %#v", c)
		}
		if l == 0 {
			continue
		}

		assert := c[0] == color.expected[0] &&
			c[1] == color.expected[1] &&
			c[2] == color.expected[2]

		if assert == false {
			t.Errorf("Invalid color schema: %#v <> %#v", color.expected, c)
		}
	}
}

func TestParseExtend(t *testing.T) {
	cases := []struct {
		value    string
		expected bimg.Extend
	}{
		{"white", bimg.ExtendWhite},
		{"black", bimg.ExtendBlack},
		{"copy", bimg.ExtendCopy},
		{"mirror", bimg.ExtendMirror},
		{"background", bimg.ExtendBackground},
		{" BACKGROUND  ", bimg.ExtendBackground},
		{"invalid", bimg.ExtendBlack},
		{"", bimg.ExtendBlack},
	}

	for _, extend := range cases {
		c := parseExtendMode(extend.value)
		if c != extend.expected {
			t.Errorf("Invalid extend value : %d != %d", c, extend.expected)
		}
	}
}

func TestGravity(t *testing.T) {
	cases := []struct {
		gravityValue   string
		smartCropValue bool
	}{
		{gravityValue: "foo", smartCropValue: false},
		{gravityValue: "smart", smartCropValue: true},
	}

	for _, td := range cases {
		str := "g=" + td.gravityValue
		io := readParams(str)
		if (io.Gravity == Gravity9Smart) != td.smartCropValue {
			t.Errorf("Expected %t to be %t, test data: %+v", io.Gravity == Gravity9Smart, td.smartCropValue, td)
		}
	}
}

func TestReadMapParams(t *testing.T) {
	cases := []struct {
		params   map[string]interface{}
		expected ImageOptions
	}{
		{
			map[string]interface{}{
				"w":    100,
				"lo":   0.1,
				"f":    "webp",
				"mono": true,
				"g":    "4",
				"b":    "ffc896",
			},
			ImageOptions{
				Width:          100,
				OverlayOpacity: 0.1,
				OutputFormat:   "webp",
				Monochrome:     true,
				Gravity:        Gravity9MiddleLeft,
				Background:     []uint8{255, 200, 150},
			},
		},
	}

	for _, test := range cases {
		opts := readMapParams(test.params)
		if opts.Width != test.expected.Width {
			t.Errorf("Invalid width: %d != %d", opts.Width, test.expected.Width)
		}
		if opts.OverlayOpacity != test.expected.OverlayOpacity {
			t.Errorf("Invalid overlay opacity: %#v != %#v", opts.OverlayOpacity, test.expected.OverlayOpacity)
		}
		if opts.OutputFormat != test.expected.OutputFormat {
			t.Errorf("Invalid output format: %s != %s", opts.OutputFormat, test.expected.OutputFormat)
		}
		if opts.Monochrome != test.expected.Monochrome {
			t.Errorf("Invalid monochrome: %#v != %#v", opts.Monochrome, test.expected.Monochrome)
		}
		if opts.Gravity != test.expected.Gravity {
			t.Errorf("Invalid gravity: %#v != %#v", opts.Gravity, test.expected.Gravity)
		}
		if opts.Background[0] != test.expected.Background[0] || opts.Background[1] != test.expected.Background[1] || opts.Background[2] != test.expected.Background[2] {
			t.Errorf("Invalid background: %#v != %#v", opts.Background, test.expected.Background)
		}
	}
}

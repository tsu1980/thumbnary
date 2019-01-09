package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/h2non/bimg.v1"
)

var allowedParams = map[string]string{
	"w": "int",
	"h": "int",
	"u": "bool",
	"m": "resizemode",
	//	"c":      "rectInt",
	//	"cr":     "rectFloat",
	"g": "gravity9",
	"b": "hexcolor",

	"l":  "string",
	"lx": "int",
	"ly": "int",
	"lg": "gravity9",
	"lo": "float",

	"mono": "bool",

	"f": "string",
	"q": "int",
}

func validateImageOptions(opts ImageOptions, o ServerOptions) error {
	outputArea := opts.Width * opts.Height
	if o.MaxOutputMP > 0 && outputArea > (o.MaxOutputMP*1000000) {
		return fmt.Errorf("The output image area(%dx%d) is exceed maximum area(%dMP)", opts.Width, opts.Height, o.MaxOutputMP)
	}
	return nil
}

func readParams(inputParamsStr string) ImageOptions {
	if inputParamsStr == "none" {
		return ImageOptionsNoConvert
	}

	paramsMap := make(map[string]string)

	r := regexp.MustCompile(`([^,=]+)=([^,=]+)`)
	paramsList := r.FindAllStringSubmatch(inputParamsStr, -1)
	for i := 0; i < len(paramsList); i++ {
		paramsMap[paramsList[i][1]] = paramsList[i][2]
	}

	params := make(map[string]interface{})

	for key, kind := range allowedParams {
		param := paramsMap[key]
		params[key] = parseParam(param, kind)
	}

	opts := mapImageParams(params)
	return opts
}

func readMapParams(options map[string]interface{}) ImageOptions {
	params := make(map[string]interface{})

	for key, kind := range allowedParams {
		value, ok := options[key]
		if !ok {
			// Force type defaults
			params[key] = parseParam("", kind)
			continue
		}

		// Parse non JSON primitive types that would be represented as string types
		if kind == "color" || kind == "hexcolor" || kind == "colorspace" || kind == "gravity" || kind == "gravity9" || kind == "extend" {
			if v, ok := value.(string); ok {
				params[key] = parseParam(v, kind)
			}
		} else if kind == "int" {
			if v, ok := value.(float64); ok {
				params[key] = int(v)
			}
			if v, ok := value.(int); ok {
				params[key] = v
			}
		} else {
			params[key] = value
		}
	}

	return mapImageParams(params)
}

func parseParam(param, kind string) interface{} {
	if kind == "int" {
		return parseInt(param)
	}
	if kind == "float" {
		return parseFloat(param)
	}
	if kind == "color" {
		return parseColor(param)
	}
	if kind == "colorspace" {
		return parseColorspace(param)
	}
	if kind == "gravity" {
		return parseGravity(param)
	}
	if kind == "bool" {
		return parseBool(param)
	}
	if kind == "extend" {
		return parseExtendMode(param)
	}
	if kind == "hexcolor" {
		return parseHexColor(param)
	}
	if kind == "gravity9" {
		return parseGravity9(param)
	}
	if kind == "rectInt" {
		return parseRectInt(param)
	}
	if kind == "rectFloat" {
		return parseRectFloat(param)
	}
	if kind == "resizemode" {
		return parseResizeMode(param)
	}
	return param
}

func mapImageParams(params map[string]interface{}) ImageOptions {
	return ImageOptions{
		Width:          params["w"].(int),
		Height:         params["h"].(int),
		Upscale:        params["u"].(bool),
		ResizeMode:     params["m"].(ResizeMode),
		Gravity:        params["g"].(Gravity9),
		Background:     params["b"].([]uint8),
		OverlayURL:     params["l"].(string),
		OverlayX:       params["lx"].(int),
		OverlayY:       params["ly"].(int),
		OverlayGravity: params["lg"].(Gravity9),
		OverlayOpacity: float32(params["lo"].(float64)),
		Monochrome:     params["mono"].(bool),
		OutputFormat:   params["f"].(string),
		Quality:        params["q"].(int),
	}
}

func parseBool(val string) bool {
	value, _ := strconv.ParseBool(val)
	return value
}

func parseInt(param string) int {
	return int(math.Floor(parseFloat(param) + 0.5))
}

func parseFloat(param string) float64 {
	val, _ := strconv.ParseFloat(param, 64)
	return math.Abs(val)
}

func parseColorspace(val string) bimg.Interpretation {
	if val == "bw" {
		return bimg.InterpretationBW
	}
	return bimg.InterpretationSRGB
}

func parseColor(val string) []uint8 {
	const max float64 = 255
	buf := []uint8{}
	if val != "" {
		for _, num := range strings.Split(val, ",") {
			n, _ := strconv.ParseUint(strings.Trim(num, " "), 10, 8)
			buf = append(buf, uint8(math.Min(float64(n), max)))
		}
	}
	return buf
}

func parseHexColor(val string) []uint8 {
	var r, g, b uint8

	if len(val) == 3 {
		fmt.Sscanf(val, "%1x%1x%1x", &r, &g, &b)
		r *= 17
		g *= 17
		b *= 17
	} else {
		fmt.Sscanf(val, "%02x%02x%02x", &r, &g, &b)
	}
	return []uint8{r, g, b}
}

func parseRectInt(val string) []int {
	var x1, y1, x2, y2 int

	if n, _ := fmt.Sscanf(val, "%d,%d,%d,%d", &x1, &y1, &x2, &y2); n == 4 {
		return []int{x1, y1, x2, y2}
	}
	return []int{x1, y1, x2, y2}
}

func parseRectFloat(val string) []float32 {
	var x1, y1, x2, y2 float32

	if n, _ := fmt.Sscanf(val, "%f,%f,%f,%f", &x1, &y1, &x2, &y2); n == 4 {
		return []float32{x1, y1, x2, y2}
	}
	return []float32{x1, y1, x2, y2}
}

func parseExtendMode(val string) bimg.Extend {
	val = strings.TrimSpace(strings.ToLower(val))
	if val == "white" {
		return bimg.ExtendWhite
	}
	if val == "copy" {
		return bimg.ExtendCopy
	}
	if val == "mirror" {
		return bimg.ExtendMirror
	}
	if val == "background" {
		return bimg.ExtendBackground
	}
	return bimg.ExtendBlack
}

func parseGravity(val string) bimg.Gravity {
	var m = map[string]bimg.Gravity{
		"south": bimg.GravitySouth,
		"north": bimg.GravityNorth,
		"east":  bimg.GravityEast,
		"west":  bimg.GravityWest,
		"smart": bimg.GravitySmart,
	}

	val = strings.TrimSpace(strings.ToLower(val))
	if g, ok := m[val]; ok {
		return g
	}

	return bimg.GravityCentre
}

func parseGravity9(val string) Gravity9 {
	var m = map[string]Gravity9{
		"1":     Gravity9TopLeft,
		"2":     Gravity9TopCenter,
		"3":     Gravity9TopRight,
		"4":     Gravity9MiddleLeft,
		"5":     Gravity9MiddleCenter,
		"6":     Gravity9MiddleRight,
		"7":     Gravity9BottomLeft,
		"8":     Gravity9BottomCenter,
		"9":     Gravity9BottomRight,
		"smart": Gravity9Smart,
	}

	val = strings.TrimSpace(strings.ToLower(val))
	if g, ok := m[val]; ok {
		return g
	}

	return Gravity9MiddleCenter
}

func parseResizeMode(val string) ResizeMode {
	var m = map[string]ResizeMode{
		"scale": ResizeModeScale,
		"crop":  ResizeModeCrop,
		"fit":   ResizeModeFit,
		"pad":   ResizeModePad,
	}

	val = strings.TrimSpace(strings.ToLower(val))
	if a, ok := m[val]; ok {
		return a
	}

	return ResizeModeCrop
}

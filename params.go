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
	"width":       "int",
	"height":      "int",
	"quality":     "int",
	"top":         "int",
	"left":        "int",
	"areawidth":   "int",
	"areaheight":  "int",
	"compression": "int",
	"rotate":      "int",
	"margin":      "int",
	"factor":      "int",
	"dpi":         "int",
	"textwidth":   "int",
	"opacity":     "float",
	"flip":        "bool",
	"flop":        "bool",
	"nocrop":      "bool",
	"noprofile":   "bool",
	"norotation":  "bool",
	"noreplicate": "bool",
	"force":       "bool",
	"embed":       "bool",
	"stripmeta":   "bool",
	"text":        "string",
	"font":        "string",
	"type":        "string",
	"color":       "color",
	"colorspace":  "colorspace",
	"gravity":     "gravity",
	"background":  "color",
	"extend":      "extend",
	"sigma":       "float",
	"minampl":     "float",
	"operations":  "json",

	// for new url format
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

func readParams(inputParamsStr string) ImageOptions {
	r := regexp.MustCompile(`([^,=]+)=([^,=]+)`)
	paramsList := r.FindAllStringSubmatch(inputParamsStr, -1)
	paramsMap := make(map[string]string)
	for i := 0; i < len(paramsList); i++ {
		paramsMap[paramsList[i][1]] = paramsList[i][2]
	}

	params := make(map[string]interface{})

	for key, kind := range allowedParams {
		param := paramsMap[key]
		params[key] = parseParam(param, kind)
	}

	opts := mapImageParams(params)
	opts.Type = opts.NewFileType
	if opts.NewMonochrome {
		opts.Colorspace = bimg.InterpretationBW
	}
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
		if kind == "color" || kind == "colorspace" || kind == "gravity" || kind == "extend" {
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
		Width:         params["width"].(int),
		Height:        params["height"].(int),
		Top:           params["top"].(int),
		Left:          params["left"].(int),
		AreaWidth:     params["areawidth"].(int),
		AreaHeight:    params["areaheight"].(int),
		DPI:           params["dpi"].(int),
		Quality:       params["quality"].(int),
		TextWidth:     params["textwidth"].(int),
		Compression:   params["compression"].(int),
		Rotate:        params["rotate"].(int),
		Factor:        params["factor"].(int),
		Color:         params["color"].([]uint8),
		Text:          params["text"].(string),
		Font:          params["font"].(string),
		Type:          params["type"].(string),
		Flip:          params["flip"].(bool),
		Flop:          params["flop"].(bool),
		Embed:         params["embed"].(bool),
		NoCrop:        params["nocrop"].(bool),
		Force:         params["force"].(bool),
		NoReplicate:   params["noreplicate"].(bool),
		NoRotation:    params["norotation"].(bool),
		NoProfile:     params["noprofile"].(bool),
		StripMetadata: params["stripmeta"].(bool),
		Opacity:       float32(params["opacity"].(float64)),
		Extend:        params["extend"].(bimg.Extend),
		Gravity:       params["gravity"].(bimg.Gravity),
		Colorspace:    params["colorspace"].(bimg.Interpretation),
		Background:    params["background"].([]uint8),
		Sigma:         params["sigma"].(float64),
		MinAmpl:       params["minampl"].(float64),

		NewWidth:          params["w"].(int),
		NewHeight:         params["h"].(int),
		NewUpscale:        params["u"].(bool),
		NewResizeMode:     params["m"].(ResizeMode),
		NewGravity:        params["g"].(Gravity9),
		NewBackground:     params["b"].([]uint8),
		NewOverlayURL:     params["l"].(string),
		NewOverlayX:       params["lx"].(int),
		NewOverlayY:       params["ly"].(int),
		NewOverlayGravity: params["lg"].(Gravity9),
		NewOverlayOpacity: float32(params["lo"].(float64)),
		NewMonochrome:     params["mono"].(bool),
		NewFileType:       params["f"].(string),
		NewQuality:        params["q"].(int),
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

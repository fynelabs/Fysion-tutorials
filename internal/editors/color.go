package editors

import (
	"fmt"
	"image/color"
)

func colorForHex(s string) color.Color {
	var rgb int
	_, err := fmt.Sscanf(s, "#%x", &rgb)
	if err != nil {
		return color.Transparent
	}

	b := rgb & 0xff
	gg := rgb >> 8 & 0xff
	r := rgb >> 16 & 0xff
	return color.NRGBA{R: uint8(r), G: uint8(gg), B: uint8(b), A: 0xff}
}

func hexForColor(c color.Color) string {
	ch := color.RGBAModel.Convert(c).(color.RGBA)
	return fmt.Sprintf("#%.2x%.2x%.2x", ch.R, ch.G, ch.B)
}

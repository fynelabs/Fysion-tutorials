package editors

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorForHex(t *testing.T) {
	assert.True(t, colorsMatch(colorForHex("#000000"), color.Black))
	assert.True(t, colorsMatch(colorForHex("#ffffff"), color.White))
	assert.True(t, colorsMatch(colorForHex("#FFffFF"), color.White))
	assert.True(t, colorsMatch(colorForHex("#c0c0c0"), color.Gray{0xc0}))

	assert.False(t, colorsMatch(colorForHex("#000001"), color.Black))
	assert.False(t, colorsMatch(colorForHex("#fffeff"), color.White))
	assert.False(t, colorsMatch(colorForHex("#FEffFF"), color.White))
}

func TestHexForColor(t *testing.T) {
	assert.Equal(t, "#000000", hexForColor(color.Black))
	assert.Equal(t, "#ffffff", hexForColor(color.White))
	assert.Equal(t, "#c0c0c0", hexForColor(color.Gray{0xc0}))

	assert.NotEqual(t, "#000001", hexForColor(color.Black))
	assert.NotEqual(t, "#fffeff", hexForColor(color.White))
	assert.NotEqual(t, "#FEffFF", hexForColor(color.White))
}

func colorsMatch(a, b color.Color) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

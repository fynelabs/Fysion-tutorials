//go:generate fyne bundle -o bundled.go assets

package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type fysionTheme struct {
	fyne.Theme
}

func newFysionTheme() fyne.Theme {
	return &fysionTheme{Theme: theme.DefaultTheme()}
}

func (t *fysionTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, theme.VariantLight)
}

func (t *fysionTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Symbol || s.Monospace {
		return t.Theme.Font(s)
	}

	if s.Bold {
		if s.Italic {
			return resourcePoppinsBoldItalicTtf
		} else {
			return resourcePoppinsBoldTtf
		}
	}
	if s.Italic {
		return resourcePoppinsItalicTtf
	}
	return resourcePoppinsRegularTtf
}

func (t *fysionTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 12
	}

	return t.Theme.Size(name)
}

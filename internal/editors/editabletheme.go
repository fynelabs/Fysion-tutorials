package editors

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type editableTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant

	dark  map[fyne.ThemeColorName]color.Color
	light map[fyne.ThemeColorName]color.Color
}

func newEditableTheme() *editableTheme {
	return &editableTheme{
		Theme:   theme.DefaultTheme(),
		variant: theme.VariantLight,
		dark:    make(map[fyne.ThemeColorName]color.Color),
		light:   make(map[fyne.ThemeColorName]color.Color),
	}
}

func (e *editableTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if e.variant == theme.VariantLight {
		if c, ok := e.light[n]; ok {
			return c
		}
	} else {
		if c, ok := e.dark[n]; ok {
			return c
		}
	}
	return e.Theme.Color(n, e.variant)
}

func (e *editableTheme) setColor(n fyne.ThemeColorName, v fyne.ThemeVariant, c color.Color) {
	if v == theme.VariantLight {
		e.light[n] = c
	} else {
		e.dark[n] = c
	}
}

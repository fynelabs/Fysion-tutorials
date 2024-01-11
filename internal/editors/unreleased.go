package editors

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func setPreviewTheme(c *container.ThemeOverride, th fyne.Theme, bg *canvas.Rectangle) {
	c.Theme = th
	c.Refresh()

	bgColor := th.Color(theme.ColorNameBackground, theme.VariantDark)
	bg.FillColor = bgColor
	bg.Refresh()
}

package editors

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

func setPreviewTheme(o fyne.CanvasObject, th fyne.Theme) {
	switch c := o.(type) {
	case *fyne.Container:
		if r, ok := c.Objects[0].(*canvas.Rectangle); ok {
			r.FillColor = th.Color(theme.ColorNameBackground, theme.VariantDark)
		}
		theme.OverrideContainer(c, th)
	case fyne.Widget:
		theme.OverrideWidget(c, th)
	}
	o.Refresh()
}

package editors

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func newMobilePreview(obj fyne.CanvasObject, bg *canvas.Rectangle) fyne.CanvasObject {
	frame := canvas.NewRectangle(color.Transparent)
	frame.StrokeColor = color.Black
	frame.StrokeWidth = 6
	frame.CornerRadius = 32

	bg.CornerRadius = 32

	inset := canvas.NewRectangle(color.Black)
	inset.CornerRadius = 5
	handle := canvas.NewRectangle(color.Gray{Y: 0x66})
	handle.CornerRadius = 2

	return container.New(mobileLayout{frame: frame, bg: bg, inset: inset, handle: handle, content: obj}, bg, obj, inset, handle, frame)
}

type mobileLayout struct {
	bg, frame, inset, handle *canvas.Rectangle
	content                  fyne.CanvasObject
}

func (l mobileLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	pad := float32(4)
	l.frame.Resize(size)
	l.bg.Resize(size)

	l.content.Move(fyne.NewPos(pad, 32))
	l.content.Resize(size.SubtractWidthHeight(pad*2, 64))

	insetSize := fyne.NewSize(80, 30)
	l.inset.Move(fyne.NewPos((size.Width-insetSize.Width)/2, 0))
	l.inset.Resize(insetSize)

	handleSize := fyne.NewSize(120, 4)
	l.handle.Move(fyne.NewPos((size.Width-handleSize.Width)/2, size.Height-18))
	l.handle.Resize(handleSize)
}

func (l mobileLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(260, 480)
}

func setPreviewTheme(c *container.ThemeOverride, th fyne.Theme,
	desktopBG, mobileBG *canvas.Rectangle) {
	c.Theme = th
	c.Refresh()

	bgColor := th.Color(theme.ColorNameBackground, theme.VariantDark)
	desktopBG.FillColor = bgColor
	desktopBG.Refresh()
	mobileBG.FillColor = bgColor
	mobileBG.Refresh()
}

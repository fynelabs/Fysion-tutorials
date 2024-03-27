package editors

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type widgetSelector struct {
	widget.BaseWidget

	root, chosen fyne.CanvasObject
	cb           func(fyne.CanvasObject)
	overlay      *canvas.Rectangle
}

func newWidgetSelector(obj fyne.CanvasObject, cb func(fyne.CanvasObject)) *widgetSelector {
	overlay := canvas.NewRectangle(color.Transparent)
	overlay.StrokeWidth = 2.5

	ret := &widgetSelector{root: obj, cb: cb, overlay: overlay}
	ret.ExtendBaseWidget(ret)
	return ret
}

func (w *widgetSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewWithoutLayout(w.overlay))
}

func (w *widgetSelector) Resize(s fyne.Size) {
	w.BaseWidget.Resize(s)

	w.updateOverlay()
}

func (w *widgetSelector) Tapped(ev *fyne.PointEvent) {
	found := findChild(w.root, ev.Position)
	if found == nil {
		found = w.root
	}

	w.choose(found)
}

func (w *widgetSelector) choose(o fyne.CanvasObject) {
	w.overlay.StrokeColor = theme.PrimaryColor()
	w.overlay.Refresh()

	w.chosen = o
	w.updateOverlay()

	w.cb(o)
}

func (w *widgetSelector) updateOverlay() {
	if w.chosen == nil {
		return
	}

	pos := w.chosen.Position()
	size := w.chosen.Size()
	if w.chosen == w.root {
		pos = fyne.NewSquareOffsetPos(-theme.Padding())
		size = size.Add(fyne.NewSquareSize(theme.Padding() * 2))
	}
	w.overlay.Move(pos)
	w.overlay.Resize(size)
}

func containerOf(obj fyne.CanvasObject, root *fyne.Container) *fyne.Container {
	for _, w := range root.Objects {
		if w == obj {
			return root
		}

		switch c := w.(type) {
		case *fyne.Container:
			parent := containerOf(obj, c)
			if parent != nil {
				return parent
			}
		}
	}

	return nil
}

func findChild(obj fyne.CanvasObject, pos fyne.Position) fyne.CanvasObject {
	switch c := obj.(type) {
	case *fyne.Container:
		for _, w := range c.Objects {
			if !inside(w, pos) {
				continue
			}

			child := findChild(w, pos.Subtract(w.Position()))
			if child != nil {
				return child
			}

			return w
		}
	}

	return nil
}

func inside(o fyne.CanvasObject, p fyne.Position) bool {
	topLeft := o.Position()
	if p.X < topLeft.X || p.Y < topLeft.Y {
		return false
	}

	size := o.Size()
	return p.X < topLeft.X+size.Width && p.Y < topLeft.Y+size.Height
}

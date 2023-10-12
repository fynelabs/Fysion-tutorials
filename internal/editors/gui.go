package editors

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/defyne/pkg/gui"
)

func makeGUI(u fyne.URI) (fyne.CanvasObject, error) {
	r, err := storage.Reader(u)
	if err != nil {
		return nil, err
	}

	obj, _ := gui.DecodeJSON(r)

	// TODO get project title, from project type when we add it
	name := "Preview" // g.title.Get()
	window := container.NewInnerWindow(name, obj)
	window.CloseIntercept = func() {}

	picker := widget.NewSelect([]string{"Desktop", "iPhone 15 Max"}, func(string) {})
	picker.Selected = "Desktop"

	preview := container.NewBorder(container.NewHBox(picker), nil, nil, nil, container.NewCenter(window))
	content := container.NewStack(canvas.NewRectangle(color.Gray{Y: 0xee}),
		container.NewPadded(preview))

	return content, nil
}

package editors

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/fyne-io/defyne/pkg/gui"
)

func makeGUI(u fyne.URI) (fyne.CanvasObject, fyne.CanvasObject, error) {
	r, err := storage.Reader(u)
	if err != nil {
		return nil, nil, err
	}

	obj, _, err := gui.DecodeJSON(r)
	if err != nil {
		return nil, nil, err
	}
	bg := canvas.NewRectangle(theme.BackgroundColor())
	inner := container.NewStack(bg, container.NewPadded(obj))

	// TODO get project title, from project type when we add it
	name := "Preview" // g.title.Get()
	window := container.NewInnerWindow(name, inner)
	window.CloseIntercept = func() {}

	picker := widget.NewSelect([]string{"Desktop", "iPhone 15 Max"}, func(string) {})
	picker.Selected = "Desktop"

	preview := container.NewBorder(container.NewHBox(picker), nil, nil, nil, container.NewCenter(window))
	content := container.NewStack(canvas.NewRectangle(color.Gray{Y: 0xee}),
		container.NewPadded(preview))

	return content, makePalette(inner), nil
}

func makePalette(obj fyne.CanvasObject) fyne.CanvasObject {
	th := newEditableTheme()

	fg := newColorButton(theme.ColorNameForeground, th, func() {
		setPreviewTheme(obj, th)
	})
	bg := newColorButton(theme.ColorNameBackground, th, func() {
		setPreviewTheme(obj, th)
	})
	button := newColorButton(theme.ColorNameButton, th, func() {
		setPreviewTheme(obj, th)
	})

	var light, dark *widget.Button
	light = widget.NewButton("Light", func() {
		th.variant = theme.VariantLight
		setPreviewTheme(obj, th)
		// TODO update in a loop?
		fg.update()
		bg.update()
		button.update()

		light.Importance = widget.HighImportance
		dark.Importance = widget.MediumImportance
		light.Refresh()
		dark.Refresh()
	})
	light.Importance = widget.HighImportance
	dark = widget.NewButton("Dark", func() {
		th.variant = theme.VariantDark
		setPreviewTheme(obj, th)
		fg.update()
		bg.update()
		button.update()

		light.Importance = widget.MediumImportance
		dark.Importance = widget.HighImportance
		light.Refresh()
		dark.Refresh()
	})
	variants := container.NewGridWithColumns(2, light, dark)

	form := container.New(layout.NewFormLayout(),
		widget.NewRichTextFromMarkdown("## Brand"), layout.NewSpacer(),
		widget.NewLabel("Text"), fg,
		widget.NewLabel("Background"), bg,
		widget.NewRichTextFromMarkdown("## Widgets"), layout.NewSpacer(),
		widget.NewLabel("Button"), button,
	)

	return container.NewVBox(variants, form)
}

type colorButton struct {
	widget.BaseWidget

	name  fyne.ThemeColorName
	theme *editableTheme

	r    *canvas.Rectangle
	text *widget.Label
	fn   func()
}

func newColorButton(n fyne.ThemeColorName, th *editableTheme, fn func()) *colorButton {
	col := th.Color(n, th.variant)
	text := widget.NewLabel(hexForColor(col))
	r := canvas.NewRectangle(col)
	r.SetMinSize(fyne.NewSquareSize(text.MinSize().Height))
	b := &colorButton{r: r, text: text, name: n, theme: th, fn: fn}
	b.ExtendBaseWidget(b)
	return b
}

func (c *colorButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, c.r, nil, c.text))
}

func (c *colorButton) Tapped(_ *fyne.PointEvent) {
	dialog.ShowColorPicker("Choose Color", "Pick a Color", func(col color.Color) {
		if c == nil {
			return
		}

		c.theme.setColor(c.name, c.theme.variant, col)
		c.update()
		c.fn()
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func (c *colorButton) update() {
	c.r.FillColor = c.theme.Color(c.name, c.theme.variant)
	c.r.Refresh()
	c.text.SetText(hexForColor(c.r.FillColor))
}

func hexForColor(c color.Color) string {
	ch := color.RGBAModel.Convert(c).(color.RGBA)
	return fmt.Sprintf("#%.2x%.2x%.2x", ch.R, ch.G, ch.B)
}

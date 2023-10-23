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

func makeGUI(u fyne.URI) (Editor, error) {
	r, err := storage.Reader(u)
	if err != nil {
		return nil, err
	}

	obj, _, err := gui.DecodeJSON(r)
	if err != nil {
		return nil, err
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

	return &simpleEditor{content: content, palette: makePalette(inner)}, nil
}

func makePalette(obj fyne.CanvasObject) fyne.CanvasObject {
	th := newEditableTheme()
	form := container.New(layout.NewFormLayout())

	// use this to ask our inputs to update on theme change
	type updatable interface {
		update()
	}

	updatePreview := func() {
		setPreviewTheme(obj, th)
	}
	updateInputs := func() {
		for _, i := range form.Objects {
			if b, ok := i.(updatable); ok {
				b.update()
			}
		}
	}

	var light, dark *widget.Button
	light = widget.NewButton("Light", func() {
		th.variant = theme.VariantLight
		setPreviewTheme(obj, th)
		updateInputs()

		light.Importance = widget.HighImportance
		dark.Importance = widget.MediumImportance
		light.Refresh()
		dark.Refresh()
	})
	light.Importance = widget.HighImportance
	dark = widget.NewButton("Dark", func() {
		th.variant = theme.VariantDark
		setPreviewTheme(obj, th)
		updateInputs()

		light.Importance = widget.MediumImportance
		dark.Importance = widget.HighImportance
		light.Refresh()
		dark.Refresh()
	})
	variants := container.NewGridWithColumns(2, light, dark)

	form.Objects = []fyne.CanvasObject{
		widget.NewRichTextFromMarkdown("## Brand"), layout.NewSpacer(),
		widget.NewLabel("Foreground"), newColorButton(theme.ColorNameForeground, th, updatePreview),
		widget.NewLabel("Background"), newColorButton(theme.ColorNameBackground, th, updatePreview),
		widget.NewLabel("Highlight"), newColorButton(theme.ColorNamePrimary, th, updatePreview),

		widget.NewRichTextFromMarkdown("## Button"), layout.NewSpacer(),
		widget.NewLabel("Background"), newColorButton(theme.ColorNameButton, th, updatePreview),
		widget.NewLabel("Pressed"), newColorButton(theme.ColorNamePressed, th, updatePreview),
		widget.NewLabel("Disabled"), newColorButton(theme.ColorNameDisabledButton, th, updatePreview),

		widget.NewRichTextFromMarkdown("## Widgets"), layout.NewSpacer(),
		widget.NewLabel("Hyperlink"), newColorButton(theme.ColorNameHyperlink, th, updatePreview),
		widget.NewLabel("Header Bg"), newColorButton(theme.ColorNameHeaderBackground, th, updatePreview),
		widget.NewLabel("Input Bg"), newColorButton(theme.ColorNameInputBackground, th, updatePreview),
		widget.NewLabel("Input Border"), newColorButton(theme.ColorNameInputBorder, th, updatePreview),
		widget.NewLabel("PlaceHolder"), newColorButton(theme.ColorNamePlaceHolder, th, updatePreview),
		widget.NewLabel("ScrollBar"), newColorButton(theme.ColorNameScrollBar, th, updatePreview),
		widget.NewLabel("Separator"), newColorButton(theme.ColorNameSeparator, th, updatePreview),

		widget.NewRichTextFromMarkdown("## State"), layout.NewSpacer(),
		widget.NewLabel("Hover"), newColorButton(theme.ColorNameHover, th, updatePreview),
		widget.NewLabel("Focus"), newColorButton(theme.ColorNameFocus, th, updatePreview),
		widget.NewLabel("Selection"), newColorButton(theme.ColorNameSelection, th, updatePreview),
		widget.NewLabel("Disabled"), newColorButton(theme.ColorNameDisabled, th, updatePreview),

		widget.NewRichTextFromMarkdown("## Other"), layout.NewSpacer(),
		widget.NewLabel("Shadow"), newColorButton(theme.ColorNameShadow, th, updatePreview),
		widget.NewLabel("Menu Bg"), newColorButton(theme.ColorNameMenuBackground, th, updatePreview),
		widget.NewLabel("Overlay Bg"), newColorButton(theme.ColorNameOverlayBackground, th, updatePreview),
		widget.NewLabel("Error"), newColorButton(theme.ColorNameError, th, updatePreview),
		widget.NewLabel("Success"), newColorButton(theme.ColorNameSuccess, th, updatePreview),
		widget.NewLabel("Warning"), newColorButton(theme.ColorNameWarning, th, updatePreview),
	}

	return container.NewBorder(variants, nil, nil, nil, container.NewScroll(form))
}

type colorButton struct {
	widget.BaseWidget

	name  fyne.ThemeColorName
	theme *editableTheme

	rect *swatch
	text *widget.Entry
	fn   func()
}

func newColorButton(n fyne.ThemeColorName, th *editableTheme, fn func()) *colorButton {
	col := th.Color(n, th.variant)
	var rect *swatch

	text := widget.NewEntry()
	text.Text = hexForColor(col)
	text.OnChanged = func(s string) {
		c := colorForHex(s)

		th.setColor(n, th.variant, c)
		rect.setColor(c)
		fn()
	}

	rect = newSwatch(col, string(n), fyne.NewSquareSize(text.MinSize().Height), func(col color.Color) {
		th.setColor(n, th.variant, col)
		text.SetText(hexForColor(col))
		fn()
	})

	b := &colorButton{rect: rect, text: text, name: n, theme: th, fn: fn}
	b.ExtendBaseWidget(b)
	return b
}

func (c *colorButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, c.rect, nil, c.text))
}

func (c *colorButton) update() {
	col := c.theme.Color(c.name, c.theme.variant)
	c.rect.setColor(col)
	c.text.SetText(hexForColor(col))
}

type swatch struct {
	widget.BaseWidget

	r    *canvas.Rectangle
	fn   func(color.Color)
	name string
}

func newSwatch(c color.Color, name string, min fyne.Size, fn func(color.Color)) *swatch {
	r := canvas.NewRectangle(c)
	r.CornerRadius = theme.InputRadiusSize()
	r.SetMinSize(min)
	s := &swatch{r: r, fn: fn, name: name}
	s.ExtendBaseWidget(s)
	return s
}

func (s *swatch) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.r)
}

func (s *swatch) Tapped(_ *fyne.PointEvent) {
	title := fmt.Sprintf("Choose %s Color", s.name)
	c := dialog.NewColorPicker(title, "", func(col color.Color) {
		if col == nil {
			return
		}

		s.setColor(col)
		s.fn(col)
	}, fyne.CurrentApp().Driver().AllWindows()[0])
	c.Advanced = true
	c.Show()
}

func (s *swatch) setColor(c color.Color) {
	s.r.FillColor = c
	s.r.Refresh()
}

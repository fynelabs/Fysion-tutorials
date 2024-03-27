package editors

import (
	"errors"
	"fmt"
	"image/color"
	"strings"

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

type GUIEditor struct {
	simpleEditor

	root   fyne.CanvasObject
	tapper *widgetSelector
}

func (g *GUIEditor) RootObject() fyne.CanvasObject {
	return g.root
}

func (g *GUIEditor) SelectWidget(obj fyne.CanvasObject) {
	g.tapper.choose(obj)
}

func makeGUI(u fyne.URI) (Editor, error) {
	r, err := storage.Reader(u)
	if err != nil {
		return nil, err
	}

	defer r.Close()
	obj, _, err := gui.DecodeJSON(r)
	if err != nil {
		return nil, err
	}

	save := func() error {
		w, err := storage.Writer(u)
		if err != nil {
			return err
		}

		defer w.Close()
		return gui.EncodeJSON(obj, make(map[fyne.CanvasObject]map[string]string), w)
	}

	var tapper *widgetSelector
	th := newEditableTheme()
	themer := container.NewThemeOverride(obj, th)

	widgetNames := gui.WidgetClassList()
	toAdd := ""
	nameList := widget.NewList(
		func() int {
			return len(widgetNames)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("WidgetClass")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			class := widgetNames[id]
			name := strings.Split(class, ".")[1]
			obj.(*widget.Label).SetText(name)
		})
	nameList.OnSelected = func(id widget.ListItemID) {
		toAdd = widgetNames[id]
	}
	insert := widget.NewButton("Insert", func() {
		if toAdd == "" {
			return
		}
		if _, ok := tapper.chosen.(*fyne.Container); !ok {
			dialog.ShowError(errors.New("selected widget must be a container"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		created := gui.CreateNew(toAdd)
		tapper.chosen.(*fyne.Container).Add(created)
		themer.Refresh()
	})
	remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		if tapper.chosen == nil {
			return
		}

		root := obj.(*fyne.Container)
		c := containerOf(tapper.chosen, root)
		if c == nil {
			c = root
		}
		c.Remove(tapper.chosen)
		tapper.chosen = nil
	})
	remove.Importance = widget.DangerImportance

	widgetType := widget.NewLabel("(select widget)")
	widgetInfo := widget.NewForm(
		widget.NewFormItem("Type", widgetType),
	)

	desktopBG := canvas.NewRectangle(theme.BackgroundColor())
	tapper = newWidgetSelector(obj, func(obj fyne.CanvasObject) {
		widgetType.SetText(gui.NameOf(obj))

		items := gui.EditorFor(obj, make(map[string]string))
		widgetInfo.Items = widgetInfo.Items[:1]
		widgetInfo.Refresh()
		widgetInfo.Items = append(widgetInfo.Items, items...)
		widgetInfo.Refresh()
	})

	preview := container.NewPadded(themer, tapper)
	desktopHolder := container.NewStack(preview)

	// TODO get project title, from project type when we add it
	name := "Preview" // g.title.Get()
	window := container.NewInnerWindow(name, container.NewStack(desktopBG, preview))
	window.SetPadded(false)
	window.Move(fyne.NewPos(20, 56))
	window.CloseIntercept = func() {}

	mobileBG := canvas.NewRectangle(theme.BackgroundColor())
	mobileHolder := container.NewStack()

	desktop := container.NewMultipleWindows(window)
	mobile := container.NewCenter(newMobilePreview(mobileHolder, mobileBG))

	picker := widget.NewSelect([]string{"Desktop", "Smart Phone"}, func(mode string) {
		switch mode {
		case "Desktop":
			desktopHolder.Objects = []fyne.CanvasObject{preview}
			mobileHolder.Objects = []fyne.CanvasObject{}
			desktopHolder.Refresh()
			mobileHolder.Refresh()

			th.multiple = 1
			themer.Refresh()
			mobile.Hide()
			desktop.Show()
		default:
			desktopHolder.Objects = []fyne.CanvasObject{}
			mobileHolder.Objects = []fyne.CanvasObject{preview}
			desktopHolder.Refresh()
			mobileHolder.Refresh()

			th.multiple = 0.6
			themer.Refresh()
			mobile.Show()
			desktop.Hide()
		}
	})
	picker.Selected = "Desktop"
	mobile.Hide()

	content := container.NewStack(canvas.NewRectangle(color.Gray{Y: 0xee}), container.NewPadded(
		container.NewStack(desktop, mobile, container.NewVBox(container.NewHBox(picker)))))

	buttonRow := container.NewBorder(nil, nil, nil, remove, insert)
	addRemove := container.NewBorder(nil, buttonRow, nil, nil, nameList)
	widgetPanel := container.NewVSplit(widgetInfo, addRemove)
	widgetPanel.Offset = 0.7
	tabs := []*container.TabItem{container.NewTabItem("Theme", makeThemePalette(themer, th, desktopBG, mobileBG)),
		container.NewTabItem("Widget", widgetPanel)}
	gui := &GUIEditor{root: obj, tapper: tapper}
	gui.content = content
	gui.palettes = tabs
	gui.save = save
	return gui, nil
}

func makeThemePalette(obj *container.ThemeOverride, th *editableTheme, bg1, bg2 *canvas.Rectangle) fyne.CanvasObject {
	form := container.New(layout.NewFormLayout())

	// use this to ask our inputs to update on theme change
	type updatable interface {
		update()
	}

	updatePreview := func() {
		setPreviewTheme(obj, th, bg1, bg2)
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
		setPreviewTheme(obj, th, bg1, bg2)
		updateInputs()

		light.Importance = widget.HighImportance
		dark.Importance = widget.MediumImportance
		light.Refresh()
		dark.Refresh()
	})
	light.Importance = widget.HighImportance
	dark = widget.NewButton("Dark", func() {
		th.variant = theme.VariantDark
		setPreviewTheme(obj, th, bg1, bg2)
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

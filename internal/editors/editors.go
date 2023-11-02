package editors

import (
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

var extentions = map[string]func(fyne.URI) (Editor, error){
	".go":       makeGo,
	".gui.json": makeGUI,
	".md":       makeMarkdown,
	".png":      makeImg,
	".txt":      makeTxt,
}

var mimes = map[string]func(fyne.URI) (Editor, error){
	"text/plain": makeTxt,
}

type Editor interface {
	Content() fyne.CanvasObject
	Palettes() []*container.TabItem

	Edited() binding.Bool
	Save() error
}

func ForURI(u fyne.URI) (Editor, error) {
	name := strings.ToLower(u.Name())
	var matched func(fyne.URI) (Editor, error)
	for ext, edit := range extentions {
		pos := strings.LastIndex(name, ext)
		if pos == -1 || pos != len(name)-len(ext) {
			continue
		}

		matched = edit
		break
	}
	if matched == nil {
		edit, ok := mimes[u.MimeType()]
		if !ok {
			return nil, errors.New("unable to find editor for file: " + u.Name() + ", mime: " + u.MimeType())
		}

		return edit(u)
	}

	return matched(u)
}

func makeGo(u fyne.URI) (Editor, error) {
	// TODO code editor
	code, err := makeTxt(u)
	if code != nil {
		code.(*simpleEditor).content.(*codeEntry).TextStyle = fyne.TextStyle{Monospace: true}
	}

	return code, err
}

func makeImg(u fyne.URI) (Editor, error) {
	img := canvas.NewImageFromURI(u)
	img.FillMode = canvas.ImageFillContain
	return &simpleEditor{content: img}, nil
}

func makeMarkdown(u fyne.URI) (Editor, error) {
	code, err := makeTxt(u)
	if code == nil || err != nil {
		return nil, err
	}

	txt := code.(*simpleEditor).content.(*codeEntry)
	txt.TextStyle = fyne.TextStyle{Monospace: true}
	txt.Refresh()

	preview := widget.NewRichTextFromMarkdown(txt.Text)
	dirty := txt.OnChanged
	txt.OnChanged = func(s string) {
		preview.ParseMarkdown(s)
		dirty(s)
	}
	code.(*simpleEditor).content = container.NewHSplit(txt, container.NewScroll(preview))

	return code, err
}

type simpleEditor struct {
	content  fyne.CanvasObject
	edited   binding.Bool
	palettes []*container.TabItem

	save func() error
}

func (s *simpleEditor) Content() fyne.CanvasObject {
	return s.content
}

func (s *simpleEditor) Palettes() []*container.TabItem {
	return s.palettes
}

func (s *simpleEditor) Edited() binding.Bool {
	if s.edited == nil {
		s.edited = binding.NewBool()
	}

	return s.edited
}

func (s *simpleEditor) Save() error {
	if s.save == nil {
		return nil
	}

	err := s.save()
	if err == nil {
		s.Edited().Set(false)
	}
	return err
}

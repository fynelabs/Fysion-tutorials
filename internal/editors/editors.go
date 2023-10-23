package editors

import (
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

var extentions = map[string]func(fyne.URI) (Editor, error){
	".go":       makeGo,
	".gui.json": makeGUI,
	".md":       makeTxt,
	".png":      makeImg,
	".txt":      makeTxt,
}

var mimes = map[string]func(fyne.URI) (Editor, error){
	"text/plain": makeTxt,
}

type Editor interface {
	Content() fyne.CanvasObject
	Palette() fyne.CanvasObject

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
		code.(*simpleEditor).content.(*widget.Entry).TextStyle = fyne.TextStyle{Monospace: true}
	}

	return code, err
}

func makeImg(u fyne.URI) (Editor, error) {
	img := canvas.NewImageFromURI(u)
	img.FillMode = canvas.ImageFillContain
	return &simpleEditor{content: img}, nil
}

type simpleEditor struct {
	content, palette fyne.CanvasObject
	edited           binding.Bool

	save func() error
}

func (s *simpleEditor) Content() fyne.CanvasObject {
	return s.content
}

func (s *simpleEditor) Palette() fyne.CanvasObject {
	return s.palette
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

package editors

import (
	"errors"
	"io"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var extentions = map[string]func(fyne.URI) (fyne.CanvasObject, error){
	".go":       makeGo,
	".gui.json": makeGUI,
	".md":       makeTxt,
	".png":      makeImg,
	".txt":      makeTxt,
}

var mimes = map[string]func(fyne.URI) (fyne.CanvasObject, error){
	"text/plain": makeTxt,
}

func ForURI(u fyne.URI) (fyne.CanvasObject, error) {
	name := strings.ToLower(u.Name())
	var matched func(fyne.URI) (fyne.CanvasObject, error)
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

func makeGo(u fyne.URI) (fyne.CanvasObject, error) {
	// TODO code editor
	code, err := makeTxt(u)
	if code != nil {
		code.(*widget.Entry).TextStyle = fyne.TextStyle{Monospace: true}
	}

	return code, err
}

func makeImg(u fyne.URI) (fyne.CanvasObject, error) {
	img := canvas.NewImageFromURI(u)
	img.FillMode = canvas.ImageFillContain
	return img, nil
}

func makeTxt(u fyne.URI) (fyne.CanvasObject, error) {
	code := widget.NewEntry()

	r, err := storage.Reader(u)
	if err != nil {
		return nil, err
	}

	defer r.Close()
	data, err := io.ReadAll(r)
	code.SetText(string(data))
	return code, err
}

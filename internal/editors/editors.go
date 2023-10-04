package editors

import (
	"io"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var extentions = map[string]func(fyne.URI) fyne.CanvasObject{
	".go":  makeGo,
	".md":  makeTxt,
	".png": makeImg,
	".txt": makeTxt,
}

var mimes = map[string]func(fyne.URI) fyne.CanvasObject{
	"text/plain": makeTxt,
}

func ForURI(u fyne.URI) fyne.CanvasObject {
	ext := strings.ToLower(u.Extension())
	edit, ok := extentions[ext]
	if !ok {
		edit, ok = mimes[u.MimeType()]
		if !ok {
			return widget.NewLabel("Unable to find editor for file: " + u.Name() + ", mime: " + u.MimeType())
		}

		return edit(u)
	}

	return edit(u)
}

func makeGo(u fyne.URI) fyne.CanvasObject {
	// TODO code editor
	code := makeTxt(u)
	code.(*widget.Entry).TextStyle = fyne.TextStyle{Monospace: true}

	return code
}

func makeImg(u fyne.URI) fyne.CanvasObject {
	img := canvas.NewImageFromURI(u)
	img.FillMode = canvas.ImageFillContain
	return img
}

func makeTxt(u fyne.URI) fyne.CanvasObject {
	code := widget.NewEntry()

	r, err := storage.Reader(u)
	if err != nil {
		code.SetText("Unable to read " + u.Name())
		return code
	}

	defer r.Close()
	data, _ := io.ReadAll(r)
	code.SetText(string(data))
	return code
}

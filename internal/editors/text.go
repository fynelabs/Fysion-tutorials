package editors

import (
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type codeEntry struct {
	widget.Entry
	win fyne.Window

	save func() error
}

func newCodeEntry(w fyne.Window) *codeEntry {
	c := &codeEntry{win: w}
	c.ExtendBaseWidget(c)

	c.MultiLine = true
	return c
}

func (c *codeEntry) TypedShortcut(s fyne.Shortcut) {
	if sh, ok := s.(*desktop.CustomShortcut); ok {
		if sh.Modifier == fyne.KeyModifierShortcutDefault && sh.KeyName == fyne.KeyS {
			if c.save != nil {
				err := c.save()
				if err != nil {
					dialog.ShowError(err, c.win)
				}
			}
			return
		}
	}

	c.Entry.TypedShortcut(s)
}

func makeTxt(u fyne.URI) (Editor, error) {
	var code *codeEntry
	save := func() error {
		return saveTxt(u, code.Text)
	}
	code = newCodeEntry(fyne.CurrentApp().Driver().AllWindows()[0])

	r, err := storage.Reader(u)
	if err != nil {
		return nil, err
	}

	defer r.Close()
	data, err := io.ReadAll(r)
	code.SetText(string(data))
	edit := &simpleEditor{content: code, save: save}
	code.OnChanged = func(_ string) {
		edit.Edited().Set(true)
	}
	code.save = edit.Save

	return edit, err
}

func saveTxt(u fyne.URI, s string) error {
	w, err := storage.Writer(u)
	if err != nil {
		return err
	}

	defer w.Close()
	_, err = io.WriteString(w, s)
	return err
}

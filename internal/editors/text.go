package editors

import (
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type codeEntry struct {
	widget.Entry

	save func() error
}

func newCodeEntry(s func() error) *codeEntry {
	c := &codeEntry{save: s}
	c.ExtendBaseWidget(c)

	c.MultiLine = true
	return c
}
func (c *codeEntry) TypedShortcut(s fyne.Shortcut) {
	if sh, ok := s.(*desktop.CustomShortcut); ok {
		if sh.Modifier == fyne.KeyModifierShortcutDefault && sh.KeyName == fyne.KeyS {
			c.save()
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
	code = newCodeEntry(save)

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

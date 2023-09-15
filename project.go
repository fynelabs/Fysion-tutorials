package main

import (
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

func createProject(name string, parent fyne.ListableURI) (fyne.ListableURI, error) {
	dir, err := storage.Child(parent, name)
	if err != nil {
		return nil, err
	}

	err = storage.CreateListable(dir)
	if err != nil {
		return nil, err
	}

	mod, err := storage.Child(dir, "go.mod")
	if err != nil {
		return nil, err
	}

	w, err := storage.Writer(mod)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	_, err = io.WriteString(w, fmt.Sprintf(`module %s
	
go 1.17

require fyne.io/fyne/v2 v2.4.0
`, name))

	list, _ := storage.ListerForURI(dir)
	return list, err
}

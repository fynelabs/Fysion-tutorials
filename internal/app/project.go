package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type Project struct {
	Meta FyneApp

	Dir fyne.ListableURI
}

func NewProject(dir fyne.ListableURI) *Project {
	name := dir.Name()
	p := &Project{Dir: dir}
	p.Meta.Details.Name = name

	dataURI, err := storage.Child(dir, "FyneApp.toml")
	if err != nil {
		fyne.LogError("Failed to access toml file path", err)
		return p
	}

	meta, err := Load(dataURI)
	if err != nil {
		fyne.LogError("Failed to parse app metadata", err)
		return p
	}

	p.Meta = meta
	return p
}

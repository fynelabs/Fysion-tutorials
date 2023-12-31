package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/storage"
)

func main() {
	a := app.NewWithID("app.fysion")
	a.Settings().SetTheme(newFysionTheme())
	w := a.NewWindow("Fysion App")
	w.SetPadded(false)
	w.Resize(fyne.NewSize(1024, 768))

	ui := &gui{win: w, project: newProjectBinding()}
	w.SetContent(ui.makeGUI())
	w.SetMainMenu(ui.makeMenu(a.Preferences()))
	ui.project.AddListener(binding.NewDataListener(func() {
		p := ui.project.GetProject()
		if p != nil {
			w.SetTitle("Fysion App: " + p.Meta.Details.Name)
		}
	}))

	flag.Usage = func() {
		fmt.Println("Usage: fysion [project directory]")
	}
	flag.Parse()
	if len(flag.Args()) > 0 {
		dirPath := flag.Args()[0]
		dirPath, err := filepath.Abs(dirPath)
		if err != nil {
			fmt.Println("Error resolving project path", err)
			return
		}

		dirURI := storage.NewFileURI(dirPath)
		dir, err := storage.ListerForURI(dirURI)
		if err != nil {
			fmt.Println("Error opening project", err)
			return
		}

		ui.openProject(dir)
	} else {
		ui.showCreate(w)
	}

	w.ShowAndRun()
}

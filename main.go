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
	a := app.New()
	a.Settings().SetTheme(newFysionTheme())
	w := a.NewWindow("Fysion App")
	w.SetPadded(false)
	w.Resize(fyne.NewSize(1024, 768))

	ui := &gui{win: w, title: binding.NewString()}
	w.SetContent(ui.makeGUI())
	w.SetMainMenu(ui.makeMenu())
	ui.title.AddListener(binding.NewDataListener(func() {
		name, _ := ui.title.Get()
		w.SetTitle("Fysion App: " + name)
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
		ui.openProjectDialog()
	}

	w.ShowAndRun()
}

func (g *gui) makeMenu() *fyne.MainMenu {
	file := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Project", g.openProjectDialog),
	)

	return fyne.NewMainMenu(file)
}

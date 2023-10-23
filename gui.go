package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fysion.app/fysion/internal/dialogs"
	"fysion.app/fysion/internal/editors"
)

type gui struct {
	win   fyne.Window
	title binding.String

	fileTree binding.URITree
	content  *container.DocTabs
	openTabs map[fyne.URI]*tabItem
	palette  *fyne.Container
}

type tabItem struct {
	editor editors.Editor
	tab    *container.TabItem
}

func (g *gui) makeBanner() fyne.CanvasObject {
	title := canvas.NewText("App Creator", theme.ForegroundColor())
	title.TextSize = 14
	title.TextStyle = fyne.TextStyle{Bold: true}

	g.title.AddListener(binding.NewDataListener(func() {
		name, _ := g.title.Get()
		if name == "" {
			name = "App Creator"
		}
		title.Text = name
		title.Refresh()
	}))

	home := widget.NewButtonWithIcon("", theme.HomeIcon(), func() {})
	left := container.NewHBox(home, title)

	logo := canvas.NewImageFromResource(resourceLogoPng)
	logo.FillMode = canvas.ImageFillContain

	return container.NewStack(container.NewPadded(left), container.NewPadded(logo))
}

func (g *gui) makeGUI() fyne.CanvasObject {
	top := g.makeBanner()
	g.fileTree = binding.NewURITree()
	files := widget.NewTreeWithData(g.fileTree, func(branch bool) fyne.CanvasObject {
		return widget.NewLabel("filename.jpg")
	}, func(data binding.DataItem, branch bool, obj fyne.CanvasObject) {
		l := obj.(*widget.Label)
		u, _ := data.(binding.URI).Get()

		l.SetText(filterName(u.Name()))
	})
	files.OnSelected = func(id widget.TreeNodeID) {
		u, err := g.fileTree.GetValue(id)
		if err != nil {
			dialog.ShowError(err, g.win)
			files.Unselect(id)
			return
		}

		listable, err := storage.CanList(u)
		if listable || err != nil {
			files.Unselect(id)
			return
		}

		err = g.openFile(u)
		if err != nil {
			dialog.ShowError(err, g.win)
			files.Unselect(id)
		}
	}

	left := widget.NewAccordion(
		widget.NewAccordionItem("Files", files),
		widget.NewAccordionItem("Screens", widget.NewLabel("TODO screens")),
	)
	left.Open(0)
	left.MultiOpen = true

	rightTop := widget.NewRichTextFromMarkdown("## Settings")
	g.palette = container.NewBorder(rightTop, nil, nil, nil)

	home := widget.NewRichTextFromMarkdown(`
# Welcome to Fysion

Please open a file from the tree on the left`)

	g.content = container.NewDocTabs(
		container.NewTabItem("Home", home),
	)
	g.content.CloseIntercept = func(item *container.TabItem) {
		var u fyne.URI
		for child, childItem := range g.openTabs {
			if childItem.tab == item {
				u = child
			}
		}

		if u != nil {
			delete(g.openTabs, u)
		}
		g.content.Remove(item)
	}
	g.content.OnSelected = func(item *container.TabItem) {
		var u fyne.URI
		for child, childItem := range g.openTabs {
			if childItem.tab == item {
				u = child
				g.setPalette(childItem.editor)
			}
		}

		if u != nil {
			files.Select(u.String())
		}
	}

	dividers := [3]fyne.CanvasObject{
		widget.NewSeparator(), widget.NewSeparator(), widget.NewSeparator(),
	}
	objs := []fyne.CanvasObject{g.content, top, left, g.palette, dividers[0], dividers[1], dividers[2]}
	return container.New(newFysionLayout(top, left, g.palette, g.content, dividers), objs...)
}

func (g *gui) makeMenu() *fyne.MainMenu {
	save := fyne.NewMenuItem("Save", func() {
		current := g.content.Selected()
		for _, child := range g.openTabs {
			if child.tab != current {
				continue
			}

			err := child.editor.Save()
			if err != nil {
				dialog.ShowError(err, g.win)
			}
			break
		}
	})
	save.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierShortcutDefault}

	file := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Project", g.openProjectDialog),
		fyne.NewMenuItemSeparator(),
		save,
	)

	return fyne.NewMainMenu(file)
}

func (g *gui) openFile(u fyne.URI) error {
	if item, ok := g.openTabs[u]; ok {
		g.content.Select(item.tab)
		return nil
	}

	edit, err := editors.ForURI(u)
	if err != nil {
		return err
	}
	g.setPalette(edit)

	name := filterName(u.Name())
	item := container.NewTabItem(name, edit.Content())
	if g.openTabs == nil {
		g.openTabs = make(map[fyne.URI]*tabItem)
	}
	g.openTabs[u] = &tabItem{editor: edit, tab: item}

	dirty := edit.Edited()
	dirty.AddListener(binding.NewDataListener(func() {
		isDirty, _ := dirty.Get()
		if isDirty {
			item.Text = name + " *"
		} else {
			item.Text = name
		}
		g.content.Refresh()
	}))

	for _, tab := range g.content.Items {
		if tab.Text != name {
			continue
		}

		// fix tab
		for uri, child := range g.openTabs {
			if child.tab != tab {
				continue
			}

			parent, _ := storage.Parent(uri)
			tab.Text = parent.Name() + string([]rune{filepath.Separator}) + tab.Text
		}

		// fix item
		parent, _ := storage.Parent(u)
		item.Text = parent.Name() + string([]rune{filepath.Separator}) + item.Text
		break
	}

	g.content.Append(item)
	g.content.Select(item)

	return nil
}

func (g *gui) openProjectDialog() {
	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, g.win)
			return
		}
		if dir == nil {
			return
		}

		g.openProject(dir)
	}, g.win)
}

func (g *gui) setPalette(e editors.Editor) {
	palette := e.Palette()
	items := g.palette.Objects[:1]
	if palette != nil {
		items = append(items, palette)
	}
	g.palette.Objects = items
	g.palette.Refresh()
}

func (g *gui) showCreate(w fyne.Window) {
	var wizard *dialogs.Wizard
	intro := widget.NewLabel(`Here you can create new project!

Or open an existing one that you created earlier.`)

	open := widget.NewButton("Open Project", func() {
		wizard.Hide()
		g.openProjectDialog()
	})
	create := widget.NewButton("Create Project", func() {
		wizard.Push("Project Details", g.makeCreateDetail(wizard))
	})
	create.Importance = widget.HighImportance

	buttons := container.NewGridWithColumns(2, open, create)
	home := container.NewVBox(intro, buttons)

	wizard = dialogs.NewWizard("Create Project", home)
	wizard.Show(w)
	wizard.Resize(home.MinSize().AddWidthHeight(40, 80)) //fyne.NewSize(360, 200))
}

func (g *gui) makeCreateDetail(wizard *dialogs.Wizard) fyne.CanvasObject {
	homeDir, _ := os.UserHomeDir()
	parent := storage.NewFileURI(homeDir)
	chosen, _ := storage.ListerForURI(parent)

	name := widget.NewEntry()
	name.Validator = func(in string) error {
		if in == "" {
			return errors.New("project name is required")
		}

		return nil
	}
	var dir *widget.Button
	dir = widget.NewButton(chosen.Name(), func() {
		d := dialog.NewFolderOpen(func(l fyne.ListableURI, err error) {
			if err != nil || l == nil {
				return
			}

			chosen = l
			dir.SetText(l.Name())
		}, g.win)

		d.SetLocation(chosen)
		d.Show()
	})

	form := widget.NewForm(
		widget.NewFormItem("Name", name),
		widget.NewFormItem("Parent Directory", dir),
	)
	form.OnSubmit = func() {
		project, err := createProject(name.Text, chosen)
		if err != nil {
			dialog.ShowError(err, g.win)
			return
		}
		wizard.Hide()
		g.openProject(project)
	}

	return form
}

func filterName(name string) string {
	pos := strings.LastIndex(name, ".gui.json")
	if pos != -1 && pos == len(name)-9 {
		name = name[:len(name)-5]
	}

	return name
}

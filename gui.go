package main

import (
	"errors"
	"fmt"
	"net/url"
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

	"fysion.app/fysion/internal/app"
	"fysion.app/fysion/internal/dialogs"
	"fysion.app/fysion/internal/editors"

	xDialog "fyne.io/x/fyne/dialog"
	gui2 "github.com/fyne-io/defyne/pkg/gui"
)

type gui struct {
	win     fyne.Window
	project *projectBinding

	fileTree   binding.URITree
	screenTree binding.StringTree
	content    *container.DocTabs
	openTabs   map[string]*tabItem
	palette    *container.AppTabs
	explorer   *widget.Accordion
}

type tabItem struct {
	editor editors.Editor
	tab    *container.TabItem
}

func (g *gui) makeBanner() fyne.CanvasObject {
	title := canvas.NewText("App Creator", theme.ForegroundColor())
	title.TextSize = 14
	title.TextStyle = fyne.TextStyle{Bold: true}

	g.project.AddListener(binding.NewDataListener(func() {
		p := g.project.GetProject()
		name := "App Creator"
		if p != nil {
			name = p.Meta.Details.Name
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
	g.screenTree = binding.NewStringTree()
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

		_, err = g.openFile(u)
		if err != nil {
			dialog.ShowError(err, g.win)
			files.Unselect(id)
		}
	}

	screens := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return g.screenTree.ChildIDs(id)
		},
		func(id widget.TreeNodeID) bool {
			return len(g.screenTree.ChildIDs((id))) > 0
		},
		func(_ bool) fyne.CanvasObject {
			return widget.NewLabel("Screen Item")
		},
		func(id widget.TreeNodeID, _ bool, obj fyne.CanvasObject) {
			l := obj.(*widget.Label)
			data, _ := g.screenTree.GetValue(id)
			l.SetText(data)
		})
	screens.OnSelected = func(id widget.TreeNodeID) {
		if strings.Contains(id, "#") {
			splits := strings.Split(id, "#")
			u, _ := storage.ParseURI(splits[0])
			edit, err := g.openFile(u)
			if err != nil {
				dialog.ShowError(err, g.win)
				files.Unselect(id)
				return
			}

			pos := strings.LastIndex(splits[1], ":")
			ui, ok := edit.(*editors.GUIEditor)
			if pos == -1 || !ok {
				return
			}
			id := splits[1][pos+1:]
			obj := findObject(ui.RootObject(), id)
			if obj != nil {
				ui.SelectWidget(obj)
			}
		} else {
			u, _ := storage.ParseURI(id)
			_, err := g.openFile(u)
			if err != nil {
				dialog.ShowError(err, g.win)
				screens.Unselect(id)
			}
		}
	}
	g.screenTree.AddListener(binding.NewDataListener(screens.Refresh))
	left := widget.NewAccordion(
		widget.NewAccordionItem("Screens", screens),
		widget.NewAccordionItem("Files", files),
	)
	left.Open(0)
	left.MultiOpen = true
	g.explorer = left

	g.palette = container.NewAppTabs(
		container.NewTabItem("App", g.makeAppPalette()),
	)

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
				u, _ = storage.ParseURI(child)
			}
		}

		if u != nil {
			delete(g.openTabs, u.String())
		}
		g.content.Remove(item)
	}
	g.content.OnSelected = func(item *container.TabItem) {
		var u fyne.URI
		for child, childItem := range g.openTabs {
			if childItem.tab == item {
				u, _ = storage.ParseURI(child)
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

func (g *gui) makeAppPalette() fyne.CanvasObject {
	name := widget.NewEntry()
	id := widget.NewEntry()
	version := widget.NewEntry()

	g.project.AddListener(binding.NewDataListener(func() {
		p := g.project.GetProject()
		if p == nil {
			return
		}
		name.OnChanged = nil
		id.OnChanged = nil
		version.OnChanged = nil

		name.SetText(p.Meta.Details.Name)
		id.SetText(p.Meta.Details.ID)
		version.SetText(p.Meta.Details.Version)

		saveMeta := func(_ string) {
			metaURI, _ := storage.Child(p.Dir, "FyneApp.toml")

			data := p.Meta
			data.Details.Name = name.Text
			data.Details.ID = id.Text
			data.Details.Version = version.Text

			err := app.Save(data, metaURI)
			if err != nil {
				dialog.ShowError(err, g.win)
			} else {
				p = app.NewProject(p.Dir)
				g.project.SetProject(p)
			}
		}

		name.OnChanged = saveMeta
		id.OnChanged = saveMeta
		version.OnChanged = saveMeta
	}))

	return widget.NewForm(
		widget.NewFormItem("Name", name),
		widget.NewFormItem("ID", id),
		widget.NewFormItem("Version", version),
	)
}

func (g *gui) makeMenu(p fyne.Preferences) *fyne.MainMenu {
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

	recent := fyne.NewMenuItem("Recent Projects...", nil)
	recents := listRecents(p)
	recentItems := make([]*fyne.MenuItem, len(recents))
	for i, r := range recents {
		recentItems[i] = fyne.NewMenuItem(r.name, func() {
			g.openProject(r.dir)
		})
	}
	recent.ChildMenu = fyne.NewMenu("Recents", recentItems...)

	about := fyne.NewMenuItem("About", g.showAbout)
	file := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Project", g.openProjectDialog),
		recent,
		fyne.NewMenuItemSeparator(),
		save,
		fyne.NewMenuItemSeparator(),
		about,
	)

	return fyne.NewMainMenu(file)
}

func (g *gui) openFile(u fyne.URI) (editors.Editor, error) {
	if item, ok := g.openTabs[u.String()]; ok {
		g.content.Select(item.tab)
		return item.editor, nil
	}

	edit, err := editors.ForURI(u)
	if err != nil {
		return nil, err
	}
	g.setPalette(edit)

	if ui, ok := edit.(*editors.GUIEditor); ok {
		obj := ui.RootObject()

		g.screenTree.Remove(u.String())

		g.screenTree.Append(binding.DataTreeRootID, u.String(), u.Name()[:len(u.Name())-9])
		addObjectsToTree(obj, g.screenTree, u, u.String()+"#")
	}

	name := filterName(u.Name())
	item := container.NewTabItem(name, edit.Content())
	if g.openTabs == nil {
		g.openTabs = make(map[string]*tabItem)
	}
	g.openTabs[u.String()] = &tabItem{editor: edit, tab: item}

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

			u, _ = storage.ParseURI(uri)
			parent, _ := storage.Parent(u)
			tab.Text = parent.Name() + string([]rune{filepath.Separator}) + tab.Text
		}

		// fix item
		parent, _ := storage.Parent(u)
		item.Text = parent.Name() + string([]rune{filepath.Separator}) + item.Text
		break
	}

	g.content.Append(item)
	g.content.Select(item)

	return edit, nil
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
	palettes := e.Palettes()

	g.palette.Items = append(g.palette.Items[:1], palettes...)
	g.palette.Refresh()

	if len(g.palette.Items) > 1 {
		g.palette.SelectIndex(1)
	}
}

func (g *gui) showCreate(w fyne.Window) {
	var wizard *dialogs.Wizard
	intro := widget.NewLabel(`Here you can create new project!

Or open an existing one that you created earlier.`)

	open := widget.NewButton("Open Project", func() {
		wizard.Hide()
		g.openProjectDialog()
	})
	recent := widget.NewButton("Recent Projects", func() {
		wizard.Push("Recent Projects", g.makeRecents(wizard))
	})
	create := widget.NewButton("Create Project", func() {
		wizard.Push("Project Details", g.makeCreateDetail(wizard))
	})
	create.Importance = widget.HighImportance

	buttons := container.NewGridWithColumns(3, open, recent, create)
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

func (g *gui) makeRecents(wizard *dialogs.Wizard) fyne.CanvasObject {
	items := listRecents(fyne.CurrentApp().Preferences())
	return widget.NewList(
		func() int {
			return len(items)
		},
		func() fyne.CanvasObject {
			return widget.NewButton("Recent Project", nil)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			b := o.(*widget.Button)

			b.OnTapped = func() {
				wizard.Hide()
				g.openProject(items[i].dir)
			}
			b.SetText(items[i].name)
		},
	)
}

func filterName(name string) string {
	pos := strings.LastIndex(name, ".gui.json")
	if pos != -1 && pos == len(name)-9 {
		name = name[:len(name)-5]
	}

	return name
}

func addObjectsToTree(obj fyne.CanvasObject, tree binding.StringTree, file fyne.URI,
	root string) {
	nodeID := fmt.Sprintf(root+":%p", obj)
	nodeRoot := root
	if root[len(root)-1] == '#' {
		nodeRoot = root[:len(root)-1]
	}
	tree.Append(nodeRoot, nodeID, gui2.NameOf(obj))

	switch c := obj.(type) {
	case *fyne.Container:
		for _, o := range c.Objects {
			addObjectsToTree(o, tree, file, nodeID)
		}
	}
}

func findObject(obj fyne.CanvasObject, id string) fyne.CanvasObject {
	myID := fmt.Sprintf("%p", obj)
	if myID == id {
		return obj
	}

	switch c := obj.(type) {
	case *fyne.Container:
		for _, o := range c.Objects {
			ret := findObject(o, id)
			if ret != nil {
				return ret
			}
		}
	}

	return nil
}

func (g *gui) showAbout() {
	about, _ := url.Parse("https://fysion.app")
	sponsor, _ := url.Parse("https://fysion.app/sponsor")
	text := `A low-code UI builder using Fyne

## Sponsors

Your name here!
`

	xDialog.ShowAboutWindow(text,
		[]*widget.Hyperlink{
			widget.NewHyperlink("About", about),
			widget.NewHyperlink("Sponsor", sponsor)},
		fyne.CurrentApp())
}

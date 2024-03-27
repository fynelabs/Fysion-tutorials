package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"fysion.app/fysion/internal/app"
)

func createFile(name string, dir fyne.URI, content string, data ...interface{}) error {
	file, err := storage.Child(dir, name)
	if err != nil {
		return err
	}

	w, err := storage.Writer(file)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.WriteString(w, fmt.Sprintf(content, data...))
	return err
}

func createProject(name string, parent fyne.ListableURI) (fyne.ListableURI, error) {
	dir, err := storage.Child(parent, name)
	if err != nil {
		return nil, err
	}

	err = storage.CreateListable(dir)
	if err != nil {
		return nil, err
	}

	err = createFile("go.mod", dir, `module %s

go 1.17

require fyne.io/fyne/v2 v2.4.1
`, sanitise(name))
	if err != nil {
		return nil, err
	}

	err = createFile("FyneApp.toml", dir, `[Details]

Name = "%s"
`, name)
	if err != nil {
		return nil, err
	}

	err = createFile("main.gui.json", dir, `{
  "Type": "*fyne.Container",
  "Layout": "VBox",
  "Name": "",
  "Objects": [
    {
      "Type": "*widget.Label",
      "Name": "",
      "Struct": {
        "Hidden": false,
        "Text": "Welcome %s!",
        "Alignment": 0,
        "Wrapping": 0,
        "TextStyle": {
          "Bold": false,
          "Italic": false,
          "Monospace": false,
          "Symbol": false,
          "TabWidth": 0
        },
        "Truncation": 0,
        "Importance": 0
      }
    },
    {
      "Type": "*widget.Button",
      "Name": "",
      "Struct": {
        "Hidden": false,
        "Text": "A button",
        "Icon": null,
        "Importance": 0,
        "Alignment": 0,
        "IconPlacement": 0
      }
    }
  ]
}
`, strings.ReplaceAll(name, "\"", "\\\""))
	if err != nil {
		return nil, err
	}

	list, _ := storage.ListerForURI(dir)
	return list, err
}

func (g *gui) openProject(dir fyne.ListableURI) {
	project := app.NewProject(dir)
	g.project.Set(project)
	addRecent(project, fyne.CurrentApp().Preferences())

	// empty the data binding if we had a project loaded
	g.fileTree.Set(map[string][]string{}, map[string]fyne.URI{})

	addFilesToTree(dir, g.fileTree, g.screenTree, binding.DataTreeRootID)
	screens := g.screenTree.ChildIDs(binding.DataTreeRootID)
	if len(screens) > 0 {
		g.explorer.Items[0].Detail.(*widget.Tree).Select(screens[0])
	} else {
		g.explorer.CloseAll()
		g.explorer.Open(1)
	}
}

func addFilesToTree(dir fyne.ListableURI, tree binding.URITree, screens binding.StringTree, root string) {
	items, _ := dir.List()
	for _, uri := range items {
		name := uri.Name()
		if len(name) > 0 && (name[0] == '.' || name == "go.sum") {
			continue
		}
		pos := strings.LastIndex(name, ".gui.go")
		if pos != -1 && pos == len(name)-7 {
			continue
		}
		pos = strings.LastIndex(name, ".gui.json")
		if pos != -1 && pos == len(name)-9 {
			screens.Append(binding.DataTreeRootID, uri.String(), name[:len(name)-9])
		}

		nodeID := uri.String()
		tree.Append(root, nodeID, uri)

		isDir, err := storage.CanList(uri)
		if err != nil {
			log.Println("Failed to check for listing")
		}
		if isDir {
			child, _ := storage.ListerForURI(uri)
			addFilesToTree(child, tree, screens, nodeID)
		}
	}
}

func sanitise(in string) string {
	return strings.ReplaceAll(in, " ", "_")
}

type projectBinding struct {
	binding.Untyped
}

func newProjectBinding() *projectBinding {
	return &projectBinding{Untyped: binding.NewUntyped()}
}

func (p *projectBinding) GetProject() *app.Project {
	proj, err := p.Get()
	if proj == nil || err != nil {
		return nil
	}

	return proj.(*app.Project)
}

func (p *projectBinding) SetProject(proj *app.Project) {
	p.Set(proj)
}

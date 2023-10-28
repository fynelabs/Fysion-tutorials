package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
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
`, sanitise(name)))
	if err != nil {
		return nil, err
	}

	json, err := storage.Child(dir, "main.gui.json")
	if err != nil {
		return nil, err
	}

	w, err = storage.Writer(json)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	_, err = io.WriteString(w, fmt.Sprintf(`{
  "Object": {
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
}
`, strings.ReplaceAll(name, "\"", "\\\""))
	if err != nil {
		return nil, err
	}

	list, _ := storage.ListerForURI(dir)
	return list, err
}

func (g *gui) openProject(dir fyne.ListableURI) {
	name := dir.Name()

	g.title.Set(name)

	// empty the data binding if we had a project loaded
	g.fileTree.Set(map[string][]string{}, map[string]fyne.URI{})

	addFilesToTree(dir, g.fileTree, binding.DataTreeRootID)
}

func addFilesToTree(dir fyne.ListableURI, tree binding.URITree, root string) {
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

		nodeID := uri.String()
		tree.Append(root, nodeID, uri)

		isDir, err := storage.CanList(uri)
		if err != nil {
			log.Println("Failed to check for listing")
		}
		if isDir {
			child, _ := storage.ListerForURI(uri)
			addFilesToTree(child, tree, nodeID)
		}
	}
}

func sanitise(in string) string {
	return strings.ReplaceAll(in, " ", "_")
}

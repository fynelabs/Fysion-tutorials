package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fysion.app/fysion/internal/app"
)

type recent struct {
	name string
	dir  fyne.ListableURI
}

func addRecent(proj *app.Project, p fyne.Preferences) {
	items := listRecents(p)
	for i, item := range items {
		// don't add a duplicate
		if item.dir.String() == proj.Dir.String() {
			if i == 0 || len(items) == 1 {
				return // nothing to re-order
			}

			ordered := append([]*recent{item}, items[:i]...)
			if i < len(items)-1 {
				ordered = append(ordered, items[i+1:]...)
			}

			writeRecents(ordered, p)
			return
		}

	}

	adding := &recent{name: proj.Meta.Details.Name, dir: proj.Dir}

	all := append([]*recent{adding}, items...)
	writeRecents(all, p)
}

func listRecents(p fyne.Preferences) []*recent {
	count := p.Int("recent.count")
	ret := make([]*recent, count)

	for i := 0; i < count; i++ {
		parent := fmt.Sprintf("recent.%d.", i)
		n := p.String(parent + "name")
		uriStr := p.String(parent + "uri")
		u, err := storage.ParseURI(uriStr)
		if err != nil {
			fyne.LogError("Failed to parse recent URI", err)
			continue
		}

		dir, _ := storage.ListerForURI(u)
		adding := &recent{name: n, dir: dir}
		ret[i] = adding
	}
	return ret
}

func writeRecents(list []*recent, p fyne.Preferences) {
	p.SetInt("recent.count", len(list))

	for i, r := range list {
		parent := fmt.Sprintf("recent.%d.", i)
		p.SetString(parent+"name", r.name)
		p.SetString(parent+"uri", r.dir.String())
	}
}

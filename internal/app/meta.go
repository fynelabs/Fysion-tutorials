package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"

	"github.com/BurntSushi/toml"
)

type FyneApp struct {
	Website     string `toml:",omitempty"`
	Details     AppDetails
	Development map[string]string `toml:",omitempty"`
	Release     map[string]string `toml:",omitempty"`
}

type AppDetails struct {
	Icon     string `toml:",omitempty"`
	Name, ID string `toml:",omitempty"`
	Version  string `toml:",omitempty"`
	Build    int    `toml:",omitempty"`
}

func Load(u fyne.URI) (data FyneApp, err error) {
	r, err := storage.Reader(u)
	if err != nil {
		return data, err
	}

	defer r.Close()
	_, err = toml.NewDecoder(r).Decode(&data)
	return data, err
}

func Save(data FyneApp, u fyne.URI) error {
	w, err := storage.Writer(u)
	if err != nil {
		return err
	}

	defer w.Close()
	return toml.NewEncoder(w).Encode(&data)
}

package editors

import (
	"bytes"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/defyne/pkg/gui"
	"github.com/stretchr/testify/assert"
)

const (
	containerJSON = `{
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
          "Text": "Welcome",
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
`

	labelJSON = `{
  "Object": {
    "Type": "*widget.Label",
    "Name": "",
    "Struct": {
      "Hidden": false,
      "Text": "Welcome",
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
  }
}
`
)

func TestDecode(t *testing.T) {
	test.NewApp()
	obj, _, err := gui.DecodeJSON(strings.NewReader(labelJSON))
	assert.Nil(t, err)

	assert.NotNil(t, obj)
	l, ok := obj.(*widget.Label)
	assert.True(t, ok)
	assert.Equal(t, "Welcome", l.Text)

	test.AssertObjectRendersToImage(t, "label.png", l)
	test.AssertObjectRendersToMarkup(t, "label.xml", l)
}

func TestDecode_Container(t *testing.T) {
	test.NewApp()
	obj, _, err := gui.DecodeJSON(strings.NewReader(containerJSON))
	assert.Nil(t, err)

	assert.NotNil(t, obj)
	c, ok := obj.(*fyne.Container)
	assert.True(t, ok)
	assert.Equal(t, 2, len(c.Objects))

	l, ok := c.Objects[0].(*widget.Label)
	assert.True(t, ok)
	assert.Equal(t, "Welcome", l.Text)
	b, ok := c.Objects[1].(*widget.Button)
	assert.True(t, ok)
	assert.Equal(t, "A button", b.Text)

	test.AssertObjectRendersToImage(t, "container.png", c)
	test.AssertObjectRendersToMarkup(t, "container.xml", c)
}

func TestEncode(t *testing.T) {
	test.NewApp()
	l := widget.NewLabel("Welcome")
	w := bytes.NewBuffer(nil)

	err := gui.EncodeJSON(l, nil, w)
	assert.Nil(t, err)

	json := w.String()
	assert.NotEmpty(t, json)
	assert.Equal(t, labelJSON, json)
}

func TestEncode_Container(t *testing.T) {
	test.NewApp()
	c := container.NewVBox(widget.NewLabel("Welcome"),
		widget.NewButton("A button", func() {}))
	w := bytes.NewBuffer(nil)

	err := gui.EncodeJSON(c, nil, w)
	assert.Nil(t, err)

	json := w.String()
	assert.NotEmpty(t, json)
	assert.Equal(t, containerJSON, json)
}

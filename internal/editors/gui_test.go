package editors

import (
	"bytes"
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/defyne/pkg/gui"
	"github.com/stretchr/testify/assert"
)

const labelJSON = `{
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
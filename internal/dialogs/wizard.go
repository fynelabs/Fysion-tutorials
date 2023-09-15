package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Wizard struct {
	title   string
	stack   []fyne.CanvasObject
	content *fyne.Container

	d dialog.Dialog
}

func NewWizard(title string, content fyne.CanvasObject) *Wizard {
	w := &Wizard{title: title, stack: []fyne.CanvasObject{content}}
	w.content = container.NewStack(content)
	return w
}

func (w *Wizard) Hide() {
	w.d.Hide()
}

func (w *Wizard) Show(win fyne.Window) {
	w.d = dialog.NewCustomWithoutButtons(w.title, w.content, win)

	w.d.Show()
}

func (w *Wizard) Pop() {
	if len(w.stack) <= 1 {
		return
	}
	w.stack = w.stack[:len(w.stack)-1]

	w.content.Objects = []fyne.CanvasObject{w.stack[len(w.stack)-1]}
	w.content.Refresh()
}

func (w *Wizard) Push(title string, content fyne.CanvasObject) {
	w.stack = append(w.stack, w.wrap(title, content))

	w.content.Objects = []fyne.CanvasObject{w.stack[len(w.stack)-1]}
	w.content.Refresh()
}

func (w *Wizard) Resize(size fyne.Size) {
	if w.d == nil {
		return
	}

	w.d.Resize(size)
}

func (w *Wizard) wrap(title string, content fyne.CanvasObject) fyne.CanvasObject {
	nav := container.NewHBox(
		widget.NewButtonWithIcon("", theme.NavigateBackIcon(), w.Pop),
		widget.NewLabel(title))

	return container.NewBorder(nav, nil, nil, nil, content)
}

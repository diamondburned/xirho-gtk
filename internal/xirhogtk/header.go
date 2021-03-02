package xirhogtk

import "github.com/gotk3/gotk3/gtk"

type headerController interface {
	PromptNewRenderer()
	SaveRenderer()
}

type header struct {
	gtk.HeaderBar
	newBtn *gtk.Button
	save   *gtk.Button
}

func newHeader(ctrl headerController) *header {
	newBtn, _ := gtk.ButtonNewFromIconName("document-open-symbolic", gtk.ICON_SIZE_BUTTON)
	newBtn.Connect("clicked", ctrl.PromptNewRenderer)
	newBtn.Show()

	save, _ := gtk.ButtonNewFromIconName("document-save-symbolic", gtk.ICON_SIZE_BUTTON)
	save.Connect("clicked", ctrl.SaveRenderer)
	save.Show()

	buttonBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	buttonBox.Add(newBtn)
	buttonBox.Add(save)
	buttonBox.Show()

	h, _ := gtk.HeaderBarNew()
	h.SetTitle("xirho-gtk")
	h.SetShowCloseButton(true)
	h.PackStart(buttonBox)

	return &header{
		HeaderBar: *h,
		newBtn:    newBtn,
		save:      save,
	}
}

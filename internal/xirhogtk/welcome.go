package xirhogtk

import "github.com/gotk3/gotk3/gtk"

type welcome struct {
	gtk.Frame
}

const fileDialogResponse = gtk.RESPONSE_ACCEPT // OK response

func newWelcome(parent gtk.IWindow, loader NewRendererSetter) *welcome {
	// File Chooser

	fileFilter, _ := gtk.FileFilterNew()
	fileFilter.AddPattern("*.xml")  // Flame XML
	fileFilter.AddPattern("*.json") // JSON

	fileDlg, _ := gtk.FileChooserDialogNewWith1Button(
		"Import a Flame/JSON file",
		parent, gtk.FILE_CHOOSER_ACTION_OPEN,
		"Import", fileDialogResponse,
	)
	fileDlg.SetSelectMultiple(false)
	fileDlg.AddFilter(fileFilter)

	fileBtn, _ := gtk.FileChooserButtonNewWithDialog(fileDlg)
	fileBtn.Show()

	// Resolution Box

	width, _ := gtk.SpinButtonNew(nil, 1, 0)
	width.SetRange(1, 9999)
	width.SetValue(1024)
	width.Show()

	xText, _ := gtk.LabelNew(`<span size="x-large">x</span>`)
	xText.SetUseMarkup(true)
	xText.Show()

	height, _ := gtk.SpinButtonNew(nil, 1, 0)
	height.SetRange(1, 9999)
	height.SetValue(1024)
	height.Show()

	// Box

	fileText, _ := gtk.LabelNew("Import a Flame/JSON file")
	fileText.Show()

	resText, _ := gtk.LabelNew("Canvas Resolution")
	resText.Show()

	resBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	resBox.Add(width)
	resBox.Add(xText)
	resBox.Add(height)
	resBox.Show()

	ctnue, _ := gtk.ButtonNewWithLabel("Continue")
	ctnue.SetMarginTop(10)
	ctnue.SetHAlign(gtk.ALIGN_CENTER)
	ctnue.Show()
	ctnue.SetSensitive(false)
	ctnue.Connect("clicked", func() {
		loader.SetNewRenderer(
			fileDlg.GetFilename(),
			int(width.GetValue()),
			int(height.GetValue()),
		)
	})

	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	box.Add(fileText)
	box.Add(fileBtn)
	box.Add(resText)
	box.Add(resBox)
	box.Add(ctnue)
	box.SetMarginBottom(10)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)
	box.Show()

	// Frame

	frame, _ := gtk.FrameNew("New Canvas")
	frame.SetHAlign(gtk.ALIGN_CENTER)
	frame.SetVAlign(gtk.ALIGN_CENTER)
	frame.Add(box)

	fileDlg.Connect("response", func(_ interface{}, respID gtk.ResponseType) {
		if respID == fileDialogResponse {
			ctnue.SetSensitive(true)
		}
	})

	return &welcome{
		Frame: *frame,
	}
}

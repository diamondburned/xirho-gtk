package xirhogtk

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/diamondburned/handy"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type sessionInfo struct {
	gtk.ScrolledWindow
	renderer *Renderer

	status *gtk.Label

	threshold  *gtk.Scale
	gamma      *gtk.SpinButton
	brightness *gtk.SpinButton

	// TODO: size and input as welcome screen
}

func newSessionInfo() *sessionInfo {
	sinfo := sessionInfo{}

	threshold, _ := gtk.ScaleNew(gtk.ORIENTATION_HORIZONTAL, nil)
	threshold.SetProperty("digits", 2)
	threshold.SetRange(0, 1)
	threshold.SetValue(0)
	threshold.SetHExpand(true)
	threshold.SetDrawValue(true)
	threshold.AddMark(0, gtk.POS_BOTTOM, "0")
	threshold.AddMark(1, gtk.POS_BOTTOM, "1")
	threshold.Show()
	threshold.Connect("value-changed", func(threshold *gtk.Scale) {
		sinfo.renderer.vplot.ToneMap.GammaMin = threshold.GetValue()
		sinfo.renderer.QueueDraw()
	})
	sinfo.threshold = threshold

	gamma, _ := gtk.SpinButtonNew(nil, 0.1, 2)
	gamma.SetRange(0, math.MaxFloat64)
	gamma.SetValue(1)
	gamma.Show()
	gamma.Connect("changed", func(gamma *gtk.SpinButton) {
		sinfo.renderer.vplot.ToneMap.Gamma = gamma.GetValue()
		sinfo.renderer.QueueDraw()
	})
	sinfo.gamma = gamma

	brightness, _ := gtk.SpinButtonNew(nil, 0.5, 2)
	brightness.SetRange(0, math.MaxFloat64)
	brightness.SetValue(1)
	brightness.Show()
	brightness.Connect("changed", func(brightness *gtk.SpinButton) {
		sinfo.renderer.vplot.ToneMap.Brightness = brightness.GetValue()
		sinfo.renderer.QueueDraw()
	})
	sinfo.brightness = brightness

	background, _ := gtk.ColorButtonNew()
	background.SetUseAlpha(true)
	background.SetRGBA(newGdkRGBA(0, 0, 0, 255))
	background.Show()
	background.Connect("color-set", func() {
		// this leaks because gotk3 lacks a marshaler for the widget.
		sinfo.renderer.SetBackgroundSolid(gdkRGBAToColor(background.GetRGBA()))
		sinfo.renderer.QueueDraw()
	})

	status, _ := gtk.LabelNew("")
	status.SetXAlign(0)
	status.Show()
	sinfo.status = status
	sinfo.UpdateInfo()

	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 15)
	box.Add(wrapPrefsGroup(
		"Status",
		status,
	))
	box.Add(wrapPrefsGroup(
		"Tone Map",
		wrapActionRow(threshold, "Threshold", ""),
		wrapActionRow(gamma, "Gamma", ""),
		wrapActionRow(brightness, "Brightness", ""),
	))
	box.Add(wrapPrefsGroup(
		"Colors",
		wrapActionRow(background, "Background Color", ""),
	))
	box.SetBorderWidth(15)
	box.Show()

	// TODO: palettes.

	scroll, _ := gtk.ScrolledWindowNew(nil, nil)
	scroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	scroll.Add(box)
	scroll.Show()

	sinfo.ScrolledWindow = *scroll
	sinfo.SetSensitive(false)

	return &sinfo
}

// UpdateInfo updates the information to the current renderer.
func (sinfo *sessionInfo) UpdateInfo() {
	if sinfo.renderer == nil {
		sinfo.status.SetText("Nothing yet...")
		return
	}

	const f = "" +
		`<b>Iterations:</b> %d` + "\n" +
		`<b>Points:</b> %d` + "\n" +
		`<b>Hit Ratio:</b> %.06f`

	iters := sinfo.renderer.xrender.Iters()
	hits := sinfo.renderer.xrender.Hits()

	sinfo.status.SetMarkup(fmt.Sprintf(
		strings.TrimSpace(f),
		iters,
		hits,
		float64(hits)/float64(iters),
	))
}

// SetRenderer sets the active renderer.
func (sinfo *sessionInfo) SetRenderer(r *Renderer) {
	sinfo.renderer = r
	sinfo.UpdateInfo()
	sinfo.SetSensitive(true)

	tmap := r.ToneMap()

	type valueSetter interface {
		GetValue() float64
		SetValue(float64)
	}

	type batchValue struct {
		value  *float64
		setter valueSetter
	}

	var batchValues = []batchValue{
		{&tmap.Gamma, sinfo.gamma},
		{&tmap.GammaMin, sinfo.threshold},
		{&tmap.Brightness, sinfo.brightness},
	}

	for _, v := range batchValues {
		if value := *v.value; value > 0 {
			v.setter.SetValue(value)
		} else {
			*v.value = v.setter.GetValue()
		}
	}
}

func gdkRGBAToColor(rgba *gdk.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(rgba.GetRed() * 255),
		G: uint8(rgba.GetGreen() * 255),
		B: uint8(rgba.GetBlue() * 255),
		A: uint8(rgba.GetAlpha() * 255),
	}
}

func newGdkRGBA(r, g, b, a uint8) *gdk.RGBA {
	return gdk.NewRGBA(
		float64(r)/0xFF,
		float64(g)/0xFF,
		float64(b)/0xFF,
		float64(a)/0xFF,
	)
}

func wrapPrefsGroup(title string, ws ...gtk.IWidget) *handy.PreferencesGroup {
	g := handy.PreferencesGroupNew()
	g.SetTitle(title)

	for _, w := range ws {
		g.Add(w)
	}

	g.Show()
	return g
}

func wrapActionRow(w gtk.IWidget, title, sub string) *handy.ActionRow {
	row := handy.ActionRowNew()
	row.Add(w)
	row.SetActivatableWidget(w)
	row.SetTitle(title)
	row.SetSubtitle(sub)
	row.Show()

	return row
}

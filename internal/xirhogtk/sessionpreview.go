package xirhogtk

import "github.com/gotk3/gotk3/gtk"

type sessionPreview struct {
	gtk.ScrolledWindow
	renderer *Renderer
}

func newSessionPreview() *sessionPreview {
	scroll, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		panic(err)
	}

	scroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scroll.SetSensitive(false)

	return &sessionPreview{
		ScrolledWindow: *scroll,
	}
}

// SetRenderer sets the visible renderer.
func (sprv *sessionPreview) SetRenderer(r *Renderer) {
	if sprv.renderer != nil {
		sprv.Remove(sprv.renderer)
	}

	sprv.renderer = r
	sprv.Add(r)
	sprv.SetSensitive(true)
}

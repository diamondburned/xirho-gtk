// Package pbimg provides a copy-free binding from Go's idiomatic image package
// to GdkPixbufs.
package pbimg

import (
	"image"
	"image/draw"
	"log"

	"github.com/gotk3/gotk3/gdk"
)

// RGBA represents a RGBA pixbuf.
type RGBA struct {
	image.RGBA
	Pixbuf *gdk.Pixbuf
}

var (
	_ image.Image = (*RGBA)(nil)
	_ draw.Image  = (*RGBA)(nil)
)

// NewRGBA creates a new RGBA image.
func NewRGBA(r image.Rectangle) *RGBA {
	pixbuf, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, r.Dx(), r.Dy())
	if err != nil {
		log.Panicln("failed to create pixbuf:", err)
	}

	// Zero out the pixels.
	pixels := []uint8(pixbuf.GetPixels())
	for i := range pixels {
		if i%4 == 3 {
			// Set all alpha pixels to black.
			pixels[i] = 255
		} else {
			pixels[i] = 0
		}
	}

	return &RGBA{
		RGBA: image.RGBA{
			Pix:    pixels,
			Stride: pixbuf.GetRowstride(),
			Rect:   image.Rect(0, 0, r.Dx(), r.Dy()),
		},
		Pixbuf: pixbuf,
	}
}

package xirhogtk

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/diamondburned/xirho-gtk/pbimg"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/pkg/errors"
	"github.com/zephyrtronium/xirho"
	"github.com/zephyrtronium/xirho/encoding"
	"github.com/zephyrtronium/xirho/encoding/flame"
	"golang.org/x/image/draw"
)

var maxProcs = runtime.GOMAXPROCS(-1)

type Renderer struct {
	gtk.Image
	scale int
	image *pbimg.RGBA

	xrender xirho.Render
	xchange chan xirho.ChangeRender
	xplot   chan xirho.PlotOnto
	ximgOut chan draw.Image

	ctx    context.Context
	cancel context.CancelFunc

	procs    int
	vplot    xirho.PlotOnto
	bgImage  image.Image
	imgRatio float64
}

// NewRenderer creates a new Renderer that's specific to the given file.
func NewRenderer(file string, w, h, oversample int) (*Renderer, error) {
	img, err := gtk.ImageNew()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create image")
	}

	// Multiply width and height by the oversample right here. Cairo will take
	// care of scaling them down.
	w *= oversample
	h *= oversample

	ctx, cancel := context.WithCancel(context.Background())

	r := Renderer{
		Image: *img,
		scale: oversample,
		image: pbimg.NewRGBA(image.Rect(0, 0, w, h)),

		xrender: xirho.Render{
			Hist:   xirho.NewHist(w, h),
			Camera: xirho.Eye(),
		},
		xchange: make(chan xirho.ChangeRender, 1),
		xplot:   make(chan xirho.PlotOnto, 1),
		ximgOut: make(chan draw.Image),

		ctx:    ctx,
		cancel: cancel,

		procs: maxProcs,
		vplot: xirho.PlotOnto{
			// TODO: swap this out with a better one.
			Scale: draw.ApproxBiLinear,
		},
		bgImage: &image.Uniform{C: color.RGBA{0, 0, 0, 255}},
	}

	go r.xrender.RenderAsync(ctx, r.xchange, r.xplot, r.ximgOut)
	r.setSize(w, h)

	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open")
	}
	defer f.Close()

	chg := xirho.ChangeRender{Procs: r.procs}

	// Lots of these code are useless and excessive, but it'll help transition
	// into drawing asynchronously in the future.

	switch ext := filepath.Ext(file); ext {
	case ".xml":
		flm, err := flame.Unmarshal(xml.NewDecoder(f))
		f.Close() // early close
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal flame xml")
		}

		chg.System = flm.System
		chg.Size = calcSize(w, h, flm.Aspect)
		chg.Camera = &flm.R.Camera
		chg.Palette = flm.R.Palette
		// r.bgImage = &image.Uniform{C: flm.BG}
		r.imgRatio = flm.Aspect
		r.vplot.ToneMap = flm.ToneMap

	case ".json":
		system, render, tm, aspect, err := encoding.Unmarshal(json.NewDecoder(f))
		f.Close() // early close
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal json")
		}

		chg.System = system
		chg.Size = calcSize(w, h, aspect)
		chg.Camera = &render.Camera
		chg.Palette = render.Palette
		r.imgRatio = aspect
		r.vplot.ToneMap = tm

	default:
		return nil, fmt.Errorf("unknown format %q", ext)
	}

	go func() {
		for recvImg := range r.ximgOut {
			// We gave it a *pbimg.RGBA so we should expect that back. We don't
			// know how to handle any other formats, so panic.
			rgba, ok := recvImg.(*pbimg.RGBA)
			if !ok {
				log.Panicf("unexpected output image of type %T", img)
			}

			// Copy the pixbuf to its own Cairo surface.
			surface, err := gdk.CairoSurfaceCreateFromPixbuf(rgba.Pixbuf, r.scale, nil)
			if err != nil {
				log.Panicln("failed to create surface from pixbuf:", err)
			}

			// This callback does not actually use any pixbufs.
			glib.IdleAdd(func() {
				img.SetFromSurface(surface)
				img.QueueDraw()
			})
		}
	}()

	select {
	case r.xchange <- chg:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return &r, nil
}

// Close always returns nil.
func (r *Renderer) Close() error {
	r.cancel()
	return nil
}

// QueueDraw queues the drawing.
func (r *Renderer) QueueDraw() {
	r.vplot.Image = r.image

	select {
	case r.xplot <- r.vplot:
	default:
		// Try taking out the old plot.
		select {
		case <-r.xplot:
		default:
		}

		// Try sending again the new plot.
		select {
		case r.xplot <- r.vplot:
		default:
		}
	}

}

// SetBackground sets the background of the output. Doing so will wipe
// everything on the pixbuf canvas.
func (r *Renderer) SetBackground(img image.Image) {
	r.bgImage = img
	r.setNewImage()
}

// SetBackgroundSolid sets the background of the output to a solid color instead
// of an image. This will wipe the canvas clean.
func (r *Renderer) SetBackgroundSolid(solid color.RGBA) {
	r.bgImage = &image.Uniform{C: solid}
	r.setNewImage()
}

func (r *Renderer) setNewImage() {
	r.setNewImageRect(r.image.Rect)
}

func (r *Renderer) setNewImageRect(rect image.Rectangle) {
	r.image = pbimg.NewRGBA(rect)
	draw.Draw(r.image, r.image.Rect, r.bgImage, image.Pt(0, 0), draw.Src)
}

// SetSize sets both the width and height of the output. The width and height
// need not match the aspect ratio.
func (r *Renderer) SetSize(w, h int) {
	w *= r.scale
	h *= r.scale

	r.setSize(w, h)
}

// setSize does SetSize without the scaling.
func (r *Renderer) setSize(w, h int) {
	r.setNewImageRect(image.Rect(0, 0, w, h))

	// Resume xirho with the new images.
	r.xchange <- xirho.ChangeRender{
		Size:  image.Pt(w, h),
		Procs: r.procs,
	}
}

// Size gets the current size.
func (r *Renderer) Size() image.Point {
	return image.Pt(r.image.Rect.Dx()/r.scale, r.image.Rect.Dy()/r.scale)
}

// AspectRatio gets the recommended aspect ratio.
func (r *Renderer) AspectRatio() float64 {
	return r.imgRatio
}

// ToneMap returns the renderer's tone map.
func (r *Renderer) ToneMap() *xirho.ToneMap {
	return &r.vplot.ToneMap
}

func calcSize(w, h int, ratio float64) image.Point {
	switch {
	case w == 0:
		return image.Pt(int(float64(w)/ratio+0.5), h)
	case h == 0:
		return image.Pt(w, int(float64(h)/ratio+0.5))
	}

	if ratio >= 1 {
		h = int(float64(w)/ratio + 0.5)
	} else {
		w = int(float64(h)/ratio + 0.5)
	}

	return image.Pt(w, h)
}

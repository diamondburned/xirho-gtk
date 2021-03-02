package xirhogtk

import (
	"fmt"
	"log"
	"time"

	"github.com/diamondburned/handy"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type Application struct {
	gtk.ApplicationWindow
	app *gtk.Application

	// window
	header *header
	view   *gtk.Stack

	// view
	welcome *welcome
	session *handy.Leaflet

	// session
	preview     *sessionPreview
	sessionInfo *sessionInfo

	// current states
	renderer     *Renderer
	renderHandle glib.SourceHandle
}

type NewRendererSetter interface {
	SetNewRenderer(file string, w, h int)
}

func New(mainApp *gtk.Application) *Application {
	app := &Application{app: mainApp}

	win, _ := gtk.ApplicationWindowNew(mainApp)
	win.SetDefaultSize(750, 500)
	app.ApplicationWindow = *win

	app.preview = newSessionPreview()
	app.preview.SetHExpand(true)
	app.preview.Show()

	app.sessionInfo = newSessionInfo()
	app.sessionInfo.SetSizeRequest(250, 50)
	app.sessionInfo.SetHExpand(false)
	app.sessionInfo.Show()

	app.session = handy.LeafletNew()
	app.session.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	app.session.Add(app.preview)
	app.session.Add(app.sessionInfo)
	app.session.Show()

	app.welcome = newWelcome(win, app)
	app.welcome.Show()

	app.view, _ = gtk.StackNew()
	app.view.SetTransitionType(gtk.STACK_TRANSITION_TYPE_CROSSFADE)
	app.view.SetTransitionDuration(75)
	app.view.AddNamed(app.welcome, "welcome")
	app.view.AddNamed(app.session, "session")
	app.view.SetVisibleChild(app.welcome)
	app.view.Show()

	app.header = newHeader(app)
	app.header.Show()

	win.Add(app.view)
	win.SetTitlebar(app.header)
	win.Connect("destroy", app.closeRenderer)

	return app
}

const (
	fps    = 10
	secsMs = uint(time.Second / time.Millisecond)
)

func (app *Application) closeRenderer() {
	if app.renderer != nil {
		app.renderer.Close()
		app.renderer = nil
	}

	if app.renderHandle > 0 {
		glib.SourceRemove(app.renderHandle)
		app.renderHandle = 0
	}
}

func (app *Application) SetNewRenderer(file string, w, h int) {
	app.closeRenderer()

	var scale = app.GetScaleFactor()
	if scale < 2 {
		scale = 2
	}

	r, err := NewRenderer(file, w, h, scale)
	if err != nil {
		// TODO: dialog
		log.Println("Error:", err)
		return
	}

	r.Show()

	app.renderer = r
	app.preview.SetRenderer(r)
	app.sessionInfo.SetRenderer(r)

	app.header.SetSubtitle(fmt.Sprintf("%s (%dx%d)", file, w, h))

	// show the session view
	app.view.SetVisibleChild(app.session)

	// redraw at given fps
	app.renderHandle = glib.TimeoutAdd(secsMs/fps, func() bool {
		r.QueueDraw()

		// TODO: move this to the draw function.
		app.sessionInfo.UpdateInfo()

		return true
	})
}

func (app *Application) SaveRenderer() {
	// TODO
}

func (app *Application) PromptNewRenderer() {
	app.header.save.SetSensitive(false)
	app.view.SetVisibleChild(app.welcome)
	app.closeRenderer()
}

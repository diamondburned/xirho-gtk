package main

import (
	"log"
	"os"
	"sync"

	"github.com/diamondburned/xirho-gtk/internal/xirhogtk"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	app, err := gtk.ApplicationNew("com.github.diamondburned.xirho-gtk", 0)
	if err != nil {
		log.Fatalln("failed to create app:", err)
	}

	app.Connect("activate", activate)

	if sig := app.Run(os.Args); sig > 0 {
		os.Exit(sig)
	}
}

var appOnce sync.Once

func activate(app *gtk.Application) {
	appOnce.Do(func() {
		xirhoApp := xirhogtk.New(app)
		xirhoApp.Show()
	})
}

package main

import (
	"embed"

	"log"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// Registered events get typed bindings, so main.js can subscribe with
	// Events.On("notice", ...) and know the payload is a string.
	application.RegisterEvent[string]("notice")
}

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
// main wires the Go backend together: one service (called from JS bindings) and
// one event channel (pushed to JS, which forwards it into Elm).
func main() {
	app := application.New(application.Options{
		Name:        "StdScout",
		Description: "Wails 3 + Elm ports + GSAP skeleton",
		Services: []application.Service{
			application.NewService(&CounterService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "StdScout",
		Width:  1100,
		Height: 700,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	// Go -> JS -> Elm: everything emitted here arrives in main.js as
	// Events.On("notice") and is then pushed into Elm through the logFromJs port.
	go func() {
		app.Event.Emit("notice", "go "+runtime.Version()+" on "+runtime.GOOS+"/"+runtime.GOARCH)

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			app.Event.Emit("notice", "heartbeat "+time.Now().Format("15:04:05"))
		}
	}()

	// Blocks until the application exits.
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

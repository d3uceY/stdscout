package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/d3uceY/stdscout/backend"
)

// The built frontend is compiled into the binary.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "StdScout",
		Description: "Every executable on your PATH",
		Services: []application.Service{
			application.NewService(&backend.Binaries{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "StdScout",
		Width:            620,
		Height:           620,
		BackgroundColour: application.NewRGB(250, 250, 250), 
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

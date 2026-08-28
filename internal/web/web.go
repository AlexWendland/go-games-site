package web

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/AlexWendland/go-games-site/ui"
)

func WebServerHandler() http.Handler {
	sub, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		log.Fatal("dist directory not found in embedded filesystem")
	}
	return http.FileServer(http.FS(sub))
}

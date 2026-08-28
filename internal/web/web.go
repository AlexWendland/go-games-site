package web

import (
	"io/fs"
	"net/http"
	"log"

	"github.com/AlexWendland/go-games-site/ui"
)

func NewHandler() http.Handler {
	sub, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		log.Fatal("dist directory not found in embedded filesystem")
	}
	return http.FileServer(http.FS(sub))
}

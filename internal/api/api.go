package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AlexWendland/go-games-site/internal/db"
)

type SimpleResponse struct {
	Message string
}

func hello(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("name")
	if len(name) == 0 {
		name = "Unknown"
	}

	response := SimpleResponse{
		Message: fmt.Sprintf("Hello %s", name),
	}
	json.NewEncoder(w).Encode(response)
}

func ApiHandler(dbConnection *db.DB) http.Handler {
	apiRouter := http.NewServeMux()

	apiRouter.HandleFunc("GET /hello", hello)

	return apiRouter
}

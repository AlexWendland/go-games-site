package api

import (
	"net/http"
	"encoding/json"
	"fmt"
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


func ApiHandler() http.Handler {
	apiRouter := http.NewServeMux()

	apiRouter.HandleFunc("GET /hello", hello)

	return apiRouter
}

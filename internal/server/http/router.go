package http

import "net/http"

func NewRouter(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/analyze", handler.Analyze)
	return mux
}

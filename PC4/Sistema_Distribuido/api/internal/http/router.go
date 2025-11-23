package http

import "github.com/gorilla/mux"

func NewRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods("GET")
	r.HandleFunc("/recommend/{userId}", h.Recommend).Methods("GET")
	return r
}

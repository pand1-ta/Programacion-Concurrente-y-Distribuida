package http

import "github.com/gorilla/mux"

func NewRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods("GET")
	r.HandleFunc("/users", h.GetUsers).Methods("GET")
	r.HandleFunc("/movies", h.GetMovies).Methods("GET")
	r.HandleFunc("/genres", h.GetGenres).Methods("GET")
	r.HandleFunc("/recommend/{userId}", h.Recommend).Methods("GET")

	return r
}

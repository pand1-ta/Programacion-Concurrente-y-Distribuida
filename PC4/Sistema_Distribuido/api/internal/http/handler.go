package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sdr/api/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	Service *service.RecommendationService
}

func NewHandler(s *service.RecommendationService) *Handler {
	return &Handler{Service: s}
}

func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	kQuery := r.URL.Query().Get("k")
	k := 10
	if kQuery != "" {
		if v, err := strconv.Atoi(kQuery); err == nil {
			k = v
		}
	}

	out, err := h.Service.Recommend(userId, k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

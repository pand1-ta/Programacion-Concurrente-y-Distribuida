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

// @Summary Genera recomendaciones para un usuario
// @Description Retorna una lista de películas recomendadas para el usuario especificado
// @Tags Recomendaciones
// @Param userId path int true "ID del usuario"
// @Param k query int false "Número de recomendaciones" default(10)
// @Success 200 {array} models.Movie
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Usuario no encontrado"
// @Router /recommend/{userId} [get]
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

// @Summary Verifica el estado del servicio
// @Description Devuelve 'ok' si el servicio está activo
// @Tags Salud
// @Success 200 {string} string "ok"
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

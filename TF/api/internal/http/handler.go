package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"sdr/api/internal/service"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Handler struct {
	Service *service.RecommendationService
}

func NewHandler(s *service.RecommendationService) *Handler {
	return &Handler{Service: s}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// @Summary Genera recomendaciones filtradas
// @Description Retorna películas recomendadas para un usuario, con filtros opcionales
// @Tags Recomendaciones
// @Param userId path int true "ID del usuario"
// @Param limit query int false "Cantidad de recomendaciones" default(10)
// @Param genre query string false "Género a filtrar"
// @Success 200 {array} models.Movie
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Usuario no encontrado"
// @Router /recommend/{userId} [get]
func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	limitQuery := r.URL.Query().Get("limit")
	genre := r.URL.Query().Get("genre")

	limit := 10
	if limitQuery != "" {
		if v, err := strconv.Atoi(limitQuery); err == nil {
			limit = v
		}
	}

	out, err := h.Service.Recommend(userId, limit, genre)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Try to get metrics from Redis (if available)
	var metrics map[string]interface{}
	if h.Service.Redis != nil {
		if found, _ := h.Service.Redis.GetCached(fmt.Sprintf("rec:%s:%s:%d:metrics", userId, genre, limit), &metrics); found {
			// ok
		} else {
			metrics = nil
		}
	}

	resp := map[string]any{
		"movies":  out,
		"metrics": metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RecommendWS upgrades the connection to a WebSocket and sends recommendations
// as a JSON payload. Path/query parameters are the same as the HTTP endpoint.
func (h *Handler) RecommendWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade to websocket: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer conn.Close()

	vars := mux.Vars(r)
	userId := vars["userId"]

	limitQuery := r.URL.Query().Get("limit")
	genre := r.URL.Query().Get("genre")

	limit := 10
	if limitQuery != "" {
		if v, err := strconv.Atoi(limitQuery); err == nil {
			limit = v
		}
	}

	out, err := h.Service.Recommend(userId, limit, genre)
	if err != nil {
		// send error message over WS and close
		_ = conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}
	// Try to get metrics from Redis
	var metrics map[string]interface{}
	if h.Service.Redis != nil {
		if found, _ := h.Service.Redis.GetCached(fmt.Sprintf("rec:%s:%s:%d:metrics", userId, genre, limit), &metrics); found {
			// ok
		} else {
			metrics = nil
		}
	}

	resp := map[string]any{
		"movies":  out,
		"metrics": metrics,
	}

	// enviar el objeto como JSON
	if err := conn.WriteJSON(resp); err != nil {
		return
	}
}

// @Summary Verifica el estado del servicio
// @Description Devuelve 'ok' si el servicio está activo
// @Tags Salud
// @Success 200 {string} string "ok"
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

// @Summary Lista usuarios
// @Description Devuelve la lista de usuarios con paginación
// @Tags Usuarios
// @Param page query int false "Página" default(1)
// @Param limit query int false "Límite por página" default(20)
// @Success 200 {array} string
// @Router /users [get]
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	users, err := h.Service.GetUsers(page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// @Summary Lista películas
// @Description Lista películas con filtro opcional por género y paginación
// @Tags Películas
// @Param genre query string false "Género"
// @Param page query int false "Página" default(1)
// @Param limit query int false "Límite por página" default(20)
// @Success 200 {array} models.Movie
// @Router /movies [get]
func (h *Handler) GetMovies(w http.ResponseWriter, r *http.Request) {
	genre := r.URL.Query().Get("genre")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	movies, err := h.Service.GetMovies(genre, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// @Summary Lista de géneros disponibles
// @Description Devuelve todos los géneros únicos encontrados en las películas
// @Tags Géneros
// @Success 200 {array} string
// @Router /genres [get]
func (h *Handler) GetGenres(w http.ResponseWriter, r *http.Request) {
	genres := h.Service.GetGenres()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"sdr/api/internal/coordinator"
	"sdr/api/internal/data"
	"sdr/api/internal/database"
	"sdr/api/internal/models"
)

type RecommendationService struct {
	Movies   map[int]models.Movie
	Mappings *data.Mappings
	Matrix   [][]float64
	Redis    *database.RedisClient
	Mongo    *database.MongoClient
	Cluster  *coordinator.CoordinatorClient
	CacheTTL time.Duration

	Genres []string // <- géneros precargados
}

func NewRecommendationService(
	movies map[int]models.Movie,
	mappings *data.Mappings,
	matrix [][]float64,
	redis *database.RedisClient,
	mongo *database.MongoClient,
	cluster *coordinator.CoordinatorClient,
	genres []string,
) *RecommendationService {
	return &RecommendationService{
		Movies:   movies,
		Mappings: mappings,
		Matrix:   matrix,
		Redis:    redis,
		Mongo:    mongo,
		Cluster:  cluster,
		CacheTTL: time.Hour,
		Genres:   genres,
	}
}

// ---------------------------------------------------------
//    Nueva función Recommend con filtros opcionales
// ---------------------------------------------------------

func (s *RecommendationService) Recommend(userIdStr string, limit int, genre string) ([]models.Movie, error) {

	// 1. Map userIdStr → índice interno
	idx, ok := s.Mappings.UserOriginalToIndex[userIdStr]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	// Normalizar género para evitar problemas de comparación
	genre = strings.TrimSpace(strings.ToLower(genre))

	// 2. Cache key mejorado: incluye filtros
	cacheKey := fmt.Sprintf("rec:%s:%s:%d", userIdStr, genre, limit)

	var cached []models.Movie
	found, _ := s.Redis.GetCached(cacheKey, &cached)
	if found {
		return cached, nil
	}

	// 3. Pedir a los workers las recomendaciones base
	movieIdxs, err := s.Cluster.RequestRecommendations(idx, s.Matrix, limit)
	if err != nil {
		return nil, err
	}

	var results []models.Movie

	// 4. Convertir índices → Movies reales con filtro opcional
	for _, mi := range movieIdxs {

		movieIDStr := s.Mappings.MovieIndexToOriginal[mi]
		movieID, err := strconv.Atoi(movieIDStr)
		if err != nil {
			continue
		}

		mv, ok := s.Movies[movieID]
		if !ok {
			continue
		}

		// Aplicar filtro de género si corresponde
		if genre != "" {
			if !strings.Contains(strings.ToLower(mv.Genre), genre) {
				continue
			}
		}

		results = append(results, mv)

		if len(results) >= limit {
			break
		}
	}

	// 5. Cache final
	_ = s.Redis.SetCached(cacheKey, results, s.CacheTTL)

	// 6. Guardar historial en Mongo
	hist := map[string]interface{}{
		"userId": userIdStr,
		"date":   time.Now(),
		"genre":  genre,
		"limit":  limit,
		"movies": results,
	}
	_ = s.Mongo.SaveRecommendation(hist)

	return results, nil
}

func (s *RecommendationService) GetUsers(page, limit int) ([]string, error) {
	return s.Mongo.GetUsersPaginated(page, limit)
}

func (s *RecommendationService) GetMovies(genre string, page, limit int) ([]models.Movie, error) {
	genre = strings.TrimSpace(strings.ToLower(genre))
	return s.Mongo.GetMoviesPaginated(genre, page, limit)
}

func (s *RecommendationService) GetGenres() []string {
	return s.Genres
}

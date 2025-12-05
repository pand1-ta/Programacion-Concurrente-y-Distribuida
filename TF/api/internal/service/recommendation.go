package service

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"

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

	// Prepare metrics
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	// CPU times for process
	var cpuStartUser, cpuStartSystem float64
	if p, err := process.NewProcess(int32(os.Getpid())); err == nil {
		if t, err := p.Times(); err == nil {
			cpuStartUser = t.User
			cpuStartSystem = t.System
		}
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

	// finish metrics
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)
	elapsed := time.Since(start)

	// CPU end times and percentage calculation
	var cpuEndUser, cpuEndSystem float64
	if p, err := process.NewProcess(int32(os.Getpid())); err == nil {
		if t, err := p.Times(); err == nil {
			cpuEndUser = t.User
			cpuEndSystem = t.System
		}
	}

	cpuDelta := (cpuEndUser + cpuEndSystem) - (cpuStartUser + cpuStartSystem)
	elapsedSec := elapsed.Seconds()
	cpuPercent := 0.0
	cpuPercentPerCPU := 0.0
	if elapsedSec > 0 {
		// raw CPU percent (may exceed 100 if multiple CPUs used)
		cpuPercent = (cpuDelta / elapsedSec) * 100.0
		cpuPercentPerCPU = cpuPercent / float64(runtime.NumCPU())
	}

	metrics := map[string]interface{}{
		"elapsed_ms":          elapsed.Milliseconds(),
		"num_cpu":             runtime.NumCPU(),
		"num_goroutine":       runtime.NumGoroutine(),
		"cpu_user_seconds":    cpuEndUser - cpuStartUser,
		"cpu_system_seconds":  cpuEndSystem - cpuStartSystem,
		"cpu_percent":         cpuPercent,
		"cpu_percent_per_cpu": cpuPercentPerCPU,
		"mem_start_alloc":     memStart.Alloc,
		"mem_end_alloc":       memEnd.Alloc,
		"mem_total_alloc":     memEnd.TotalAlloc,
		"mem_sys":             memEnd.Sys,
	}

	// 6. Guardar historial en Mongo (incluye metrics)
	hist := map[string]interface{}{
		"userId":  userIdStr,
		"date":    time.Now(),
		"genre":   genre,
		"limit":   limit,
		"movies":  results,
		"metrics": metrics,
	}
	_ = s.Mongo.SaveRecommendation(hist)

	// Build response object: include movies + metrics so handlers can return both
	// We return the movies slice as before; handlers will call another method to fetch metrics if needed.
	// For now, embed metrics by returning a custom wrapper via a separate method.

	// To keep backward compatibility with existing callers, return movies and store metrics in Redis as additional key
	_ = s.Redis.SetCached(cacheKey+":metrics", metrics, s.CacheTTL)

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

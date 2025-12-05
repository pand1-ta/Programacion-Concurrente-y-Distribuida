package service

import (
	"fmt"
	"strconv"
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
}

func NewRecommendationService(
	movies map[int]models.Movie,
	mappings *data.Mappings,
	matrix [][]float64,
	redis *database.RedisClient,
	mongo *database.MongoClient,
	cluster *coordinator.CoordinatorClient,
) *RecommendationService {
	return &RecommendationService{
		Movies:   movies,
		Mappings: mappings,
		Matrix:   matrix,
		Redis:    redis,
		Mongo:    mongo,
		Cluster:  cluster,
		CacheTTL: time.Hour,
	}
}

func (s *RecommendationService) Recommend(userIdStr string, k int) ([]models.Movie, error) {
	// map userIdStr -> index
	idx, ok := s.Mappings.UserOriginalToIndex[userIdStr]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	cacheKey := "rec:" + userIdStr
	var cached []models.Movie
	found, _ := s.Redis.GetCached(cacheKey, &cached)
	if found {
		return cached, nil
	}

	// ask coordinator for top similar users OR to compute something
	movieIdxs, err := s.Cluster.RequestRecommendations(idx, s.Matrix, k)
	if err != nil {
		return nil, err
	}

	// movieIdxs are ordered candidate movie indices (index in matrix)
	// take top k (or top N required)
	// We'll map to MovieIDs and return titles
	var out []models.Movie
	for i, mi := range movieIdxs {
		if i >= k {
			break
		}
		movieIDStr := s.Mappings.MovieIndexToOriginal[mi]
		movieID, err := strconv.Atoi(movieIDStr)
		if err != nil {
			// fallback: produce placeholder when movieID is invalid
			out = append(out, models.Movie{MovieID: movieIDStr, Title: "Unknown"})
			continue
		}
		mv, ok := s.Movies[movieID]
		if !ok {
			// fallback: produce placeholder
			out = append(out, models.Movie{MovieID: fmt.Sprintf("%d", movieID), Title: "Unknown"})
		} else {
			out = append(out, mv)
		}
	}

	// cache
	_ = s.Redis.SetCached(cacheKey, out, s.CacheTTL)

	// persist history
	hist := map[string]interface{}{
		"userId": userIdStr,
		"date":   time.Now(),
		"movies": out,
	}
	_ = s.Mongo.SaveRecommendation(hist)

	return out, nil
}

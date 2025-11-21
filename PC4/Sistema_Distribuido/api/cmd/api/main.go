package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"sdr/api/internal/coordinator"
	"sdr/api/internal/data"
	"sdr/api/internal/database"
	httpApi "sdr/api/internal/http"
	"sdr/api/internal/models"
	"sdr/api/internal/service"
)

func main() {
	// Paths relative to working dir in container: ./dataset/...
	dsPath := "/app/dataset"

	// 1) Load CSVs
	movies, err := data.LoadMovies(dsPath + "/movies.csv")
	if err != nil {
		log.Fatalf("LoadMovies error: %v", err)
	}

	userOrigToIdx, userIdxToOrig, err := data.LoadMapping(dsPath + "/usuarios_mapping.csv")
	if err != nil {
		log.Fatalf("Load user mapping: %v", err)
	}
	movieOrigToIdx, movieIdxToOrig, err := data.LoadMapping(dsPath + "/peliculas_mapping.csv")
	if err != nil {
		log.Fatalf("Load movie mapping: %v", err)
	}

	matrixData, err := data.LoadUserMovieMatrix(dsPath + "/matriz_usuarios_peliculas.csv")
	if err != nil {
		log.Fatalf("Load matrix: %v", err)
	}

	// 2) Connect to DBs
	// Mongo
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongodb:27017"
	}
	mongoClient, err := database.NewMongoClient(mongoURI, "sdr", "history")
	if err != nil {
		log.Fatalf("Mongo connect error: %v", err)
	}
	// Seed movies & users if not present
	// Convert movies map[int]models.Movie to []models.Movie expected by SeedMoviesIfEmpty
	moviesSlice := make([]models.Movie, 0, len(movies))
	for _, m := range movies {
		moviesSlice = append(moviesSlice, m)
	}
	if err := mongoClient.SeedMoviesIfEmpty(moviesSlice); err != nil {
		log.Fatalf("Seed movies: %v", err)
	}
	if err := mongoClient.SeedUsersIfEmpty(userIdxToOrig); err != nil {
		log.Fatalf("Seed users: %v", err)
	}

	// Redis
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	if redisHost == "" {
		redisHost = "redis"
	}
	if redisPort == "" {
		redisPort = "6379"
	}
	redisClient := database.NewRedisClient(redisHost, redisPort)

	// Coordinator client (TCP)
	coordAddr := os.Getenv("COORDINATOR_ADDR")
	if coordAddr == "" {
		coordAddr = "coordinator:8081"
	}
	cluster := coordinator.NewCoordinatorClient(coordAddr)

	// Build mappings struct for service
	mappings := &data.Mappings{
		UserOriginalToIndex:  userOrigToIdx,
		UserIndexToOriginal:  userIdxToOrig,
		MovieOriginalToIndex: movieOrigToIdx,
		MovieIndexToOriginal: movieIdxToOrig,
	}

	// Service
	svc := service.NewRecommendationService(movies, mappings, matrixData.Matrix, redisClient, mongoClient, cluster)

	// HTTP
	handler := httpApi.NewHandler(svc)
	router := httpApi.NewRouter(handler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("API listening on :8080")
	log.Fatal(srv.ListenAndServe())
}

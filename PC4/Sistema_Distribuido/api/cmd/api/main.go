package main

import (
	"log"
	"net/http"
	"os"
	"time"

	_ "sdr/api/docs" // Importa la documentación generada por swag

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	"sdr/api/internal/coordinator"
	"sdr/api/internal/data"
	"sdr/api/internal/database"
	httpApi "sdr/api/internal/http"
	"sdr/api/internal/service"
)

// @title Sistema Distribuido de Recomendaciones
// @version 1.0
// @description API para generar recomendaciones personalizadas utilizando un clúster distribuido de Workers.
// @contact.name Equipo SDR
// @host localhost:8080
// @BasePath /
func main() {
	dsPath := "/app/dataset"

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

	// MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongodb:27017"
	}
	mongoClient, err := database.NewMongoClient(mongoURI, "sdr", "history")
	if err != nil {
		log.Fatalf("Mongo connect error: %v", err)
	}

	// Redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisClient := database.NewRedisClient(redisHost, redisPort)

	// Coordinador (TCP)
	coordAddr := os.Getenv("COORDINATOR_ADDR")
	if coordAddr == "" {
		coordAddr = "sdr_coordinator:8081"
	}
	cluster := coordinator.NewCoordinatorClient(coordAddr)

	mappings := &data.Mappings{
		UserOriginalToIndex:  userOrigToIdx,
		UserIndexToOriginal:  userIdxToOrig,
		MovieOriginalToIndex: movieOrigToIdx,
		MovieIndexToOriginal: movieIdxToOrig,
	}

	svc := service.NewRecommendationService(movies, mappings, matrixData.Matrix, redisClient, mongoClient, cluster)

	handler := httpApi.NewHandler(svc)
	router := mux.NewRouter()

	// Rutas principales
	router.HandleFunc("/recommend/{userId}", handler.Recommend).Methods("GET")
	router.HandleFunc("/health", handler.Health).Methods("GET")

	// Rutas Swagger
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

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

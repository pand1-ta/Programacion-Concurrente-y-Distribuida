package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============= CONFIGURACIÓN =============
const (
	MinGamesPerUser     = 3
	TopKNeighbors       = 50
	MinSimilarity       = 0.2
	NumRecommendations  = 15
	CsvFilePath         = "steam_reviews.csv"
	ConfidenceThreshold = 10
	MaxReviewsToLoad    = 5000000 // 0 para cargar todas las reseñas
	TargetUserID        = 76561198059107008
)

// ============= ESTRUCTURAS DE DATOS =============
type Review struct {
	AppID                  int
	AppName                string
	Recommended            bool
	VotesHelpful           int
	SteamPurchase          bool
	AuthorSteamID          int64
	AuthorPlaytimeAtReview int
}
type RatingMatrix struct {
	Data      map[int64]map[int]float64
	UserList  []int64
	GameList  []int
	GameNames map[int]string
}
type UserSimilarity struct {
	UserID     int64
	Similarity float64
}
type GameRecommendation struct {
	GameID    int
	GameName  string
	PredScore float64
	NumVoters int
}

type GameScore struct {
	Name   string
	Rating float64
}

// ============= FUNCIONES AUXILIARES DE PARSEO =============
func parseFlexibleInt(s string) (int, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int(f), nil
}
func parseFlexibleInt64(s string) (int64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

// ============= CARGA DE DATOS =============
func loadAllReviews(filepath string, maxLines int) []Review {
	file, err := os.Open(filepath)
	if err != nil {
		return nil
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	reviews := make([]Review, 0)
	errorCount := 0
	_, err = csvReader.Read()
	if err != nil {
		return nil
	}
	for {
		if maxLines > 0 && len(reviews) >= maxLines {
			fmt.Printf("\nLímite de %d de reseñas alcanzado.\n", maxLines)
			break
		}
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorCount++
			continue
		}
		if len(record) < 23 {
			errorCount++
			continue
		}
		appID, err1 := parseFlexibleInt(record[1])
		votesHelpful, err3 := parseFlexibleInt(record[9])
		authorSteamID, err5 := parseFlexibleInt64(record[16])
		authorPlaytimeAtReview, err6 := parseFlexibleInt(record[21])
		recommended, err2 := strconv.ParseBool(record[8])
		steamPurchase, err4 := strconv.ParseBool(record[13])
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
			errorCount++
			continue
		}
		review := Review{
			AppID: appID, AppName: record[2], Recommended: recommended,
			VotesHelpful: votesHelpful, SteamPurchase: steamPurchase, AuthorSteamID: authorSteamID,
			AuthorPlaytimeAtReview: authorPlaytimeAtReview,
		}
		reviews = append(reviews, review)
	}
	return reviews
}

// ============= CÁLCULO DE RATING =============
func calculateRating(r Review) float64 {
	var baseScore, maxScore float64
	if r.Recommended {
		baseScore, maxScore = 6.0, 10.0
	} else {
		baseScore, maxScore = 1.0, 5.0
	}
	hours := float64(r.AuthorPlaytimeAtReview) / 60.0
	playtimeModifier := math.Log(hours+1) / math.Log(200+1)
	if playtimeModifier > 1.0 {
		playtimeModifier = 1.0
	}
	rating := baseScore + (maxScore-baseScore)*playtimeModifier
	if r.VotesHelpful > 10 {
		rating += 0.3
	}
	if r.SteamPurchase && r.Recommended {
		rating += 0.2
	}
	return math.Max(1.0, math.Min(10.0, rating))
}

// ============= CONSTRUCCIÓN DE MATRIZ =============
func buildRatingMatrix(reviews []Review, minGamesPerUser int) *RatingMatrix {
	userRatings := make(map[int64]map[int]float64)
	gameNames := make(map[int]string)
	for _, review := range reviews {
		if userRatings[review.AuthorSteamID] == nil {
			userRatings[review.AuthorSteamID] = make(map[int]float64)
		}
		userRatings[review.AuthorSteamID][review.AppID] = calculateRating(review)
		gameNames[review.AppID] = review.AppName
	}
	matrix := &RatingMatrix{
		Data: make(map[int64]map[int]float64), GameNames: gameNames, UserList: make([]int64, 0),
	}
	gameSet := make(map[int]bool)
	for userID, games := range userRatings {
		if len(games) >= minGamesPerUser {
			matrix.Data[userID] = games
			matrix.UserList = append(matrix.UserList, userID)
			for gameID := range games {
				gameSet[gameID] = true
			}
		}
	}
	matrix.GameList = make([]int, 0, len(gameSet))
	for gameID := range gameSet {
		matrix.GameList = append(matrix.GameList, gameID)
	}
	return matrix
}

// ============= NORMALIZACIÓN =============
func normalizeMatrix(matrix *RatingMatrix) map[int64]map[int]float64 {
	normalized := make(map[int64]map[int]float64)
	for userID, games := range matrix.Data {
		sum := 0.0
		for _, rating := range games {
			sum += rating
		}
		userMean := sum / float64(len(games))
		variance := 0.0
		for _, rating := range games {
			diff := rating - userMean
			variance += diff * diff
		}
		stdDev := math.Sqrt(variance / float64(len(games)))
		if stdDev == 0 {
			stdDev = 1.0
		}
		normalized[userID] = make(map[int]float64)
		for gameID, rating := range games {
			normalized[userID][gameID] = (rating - userMean) / stdDev
		}
	}
	return normalized
}

// ============= SIMILITUD DE COSENO (CON PENALIZACIÓN POR CONFIANZA) =============
func calculateCosineSimilarity(normalized map[int64]map[int]float64, user1, user2 int64) float64 {
	ratings1, ok1 := normalized[user1]
	ratings2, ok2 := normalized[user2]
	if !ok1 || !ok2 {
		return 0.0
	}
	commonGames := make([]int, 0)
	for gameID := range ratings1 {
		if _, exists := ratings2[gameID]; exists {
			commonGames = append(commonGames, gameID)
		}
	}
	numCommon := len(commonGames)
	if numCommon < 2 {
		return 0.0
	}
	dotProduct, norm1, norm2 := 0.0, 0.0, 0.0
	for _, gameID := range commonGames {
		r1, r2 := ratings1[gameID], ratings2[gameID]
		dotProduct += r1 * r2
		norm1 += r1 * r1
		norm2 += r2 * r2
	}
	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}
	cosineSim := dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
	confidenceFactor := math.Log(float64(numCommon)) / math.Log(float64(ConfidenceThreshold))
	if confidenceFactor > 1.0 {
		confidenceFactor = 1.0
	}
	return cosineSim * confidenceFactor
}

// ============= USUARIOS SIMILARES - CÁLCULO SECUENCIAL Y CONCURRENTE =============

func findSimilarUsersSequential(matrix *RatingMatrix, normalized map[int64]map[int]float64, targetUser int64) []UserSimilarity {
	similarities := make([]UserSimilarity, 0)
	for _, userID := range matrix.UserList {
		if userID == targetUser {
			continue
		}
		similarity := calculateCosineSimilarity(normalized, targetUser, userID)
		if similarity > MinSimilarity {
			similarities = append(similarities, UserSimilarity{
				UserID:     userID,
				Similarity: similarity,
			})
		}
	}
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})
	return similarities
}

func findSimilarUsersConcurrent(matrix *RatingMatrix, normalized map[int64]map[int]float64, targetUser int64, numWorkers int) []UserSimilarity {
	userChannel := make(chan int64, len(matrix.UserList))
	results := make(chan UserSimilarity, len(matrix.UserList))
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for userID := range userChannel {
				if userID == targetUser {
					continue
				}
				similarity := calculateCosineSimilarity(normalized, targetUser, userID)
				if similarity > MinSimilarity {
					results <- UserSimilarity{UserID: userID, Similarity: similarity}
				}
			}
		}()
	}
	for _, userID := range matrix.UserList {
		userChannel <- userID
	}
	close(userChannel)
	wg.Wait()
	close(results)
	similarities := make([]UserSimilarity, 0, len(results))
	for result := range results {
		similarities = append(similarities, result)
	}
	sort.Slice(similarities, func(i, j int) bool { return similarities[i].Similarity > similarities[j].Similarity })
	return similarities
}

// ============= GENERACIÓN DE RECOMENDACIONES =============
func generateRecommendations(matrix *RatingMatrix, targetUser int64, similarUsers []UserSimilarity) []GameRecommendation {
	neighbors := similarUsers
	if len(neighbors) > TopKNeighbors {
		neighbors = neighbors[:TopKNeighbors]
	}
	if len(neighbors) == 0 {
		return nil
	}
	playedGames := make(map[int]bool)
	for gameID := range matrix.Data[targetUser] {
		playedGames[gameID] = true
	}
	predictions := make(map[int]float64)
	weights := make(map[int]float64)
	voters := make(map[int]int)
	for _, neighbor := range neighbors {
		for gameID, rating := range matrix.Data[neighbor.UserID] {
			if !playedGames[gameID] {
				predictions[gameID] += neighbor.Similarity * rating
				weights[gameID] += neighbor.Similarity
				voters[gameID]++
			}
		}
	}
	recommendations := make([]GameRecommendation, 0)
	for gameID, predSum := range predictions {
		if weights[gameID] > 0 && voters[gameID] >= 2 {
			predictedScore := predSum / weights[gameID]
			recommendations = append(recommendations, GameRecommendation{
				GameID: gameID, GameName: matrix.GameNames[gameID],
				PredScore: predictedScore, NumVoters: voters[gameID],
			})
		}
	}
	sort.Slice(recommendations, func(i, j int) bool { return recommendations[i].PredScore > recommendations[j].PredScore })
	return recommendations
}

// ============= GUARDAR RESULTADOS EN ARCHIVO =============
func saveResultsToFile(
	reviewsCount int,
	matrix *RatingMatrix,
	targetUser int64,
	userGames []GameScore,
	avgRating float64,
	recommendations []GameRecommendation,
	similarUsers []UserSimilarity,
	durationSequential time.Duration,
	concurrentResults map[int]time.Duration, // Recibe un mapa de resultados
) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Numero de reseñas cargadas: %d\n", reviewsCount))
	sb.WriteString(fmt.Sprintf("Usuarios en la matriz: %d | Juegos en la matriz: %d\n", len(matrix.UserList), len(matrix.GameList)))
	sb.WriteString(fmt.Sprintf("Usuario seleccionado: %d | Juegos jugados: %d | Rating promedio: %.2f\n", targetUser, len(userGames), avgRating))
	sb.WriteString("\n--- Top 5 Juegos del Usuario ---\n")
	for i := 0; i < 5 && i < len(userGames); i++ {
		sb.WriteString(fmt.Sprintf("  %d. %s (Rating: %.2f)\n", i+1, userGames[i].Name, userGames[i].Rating))
	}
	sb.WriteString("\n--- Recomendaciones Generadas ---\n")
	for i := 0; i < NumRecommendations && i < len(recommendations); i++ {
		rec := recommendations[i]
		sb.WriteString(fmt.Sprintf("%d. %s (Score Predicho: %.2f | Basado en %d vecinos)\n", i+1, rec.GameName, rec.PredScore, rec.NumVoters))
	}
	sb.WriteString("\n--- Top 5 Usuarios Similares ---\n")
	numTopUsersToShow := 5
	if len(similarUsers) < numTopUsersToShow {
		numTopUsersToShow = len(similarUsers)
	}
	for i := 0; i < numTopUsersToShow; i++ {
		simUser := similarUsers[i]
		commonGames := make([]string, 0)
		for gameID := range matrix.Data[targetUser] {
			if _, exists := matrix.Data[simUser.UserID][gameID]; exists {
				commonGames = append(commonGames, matrix.GameNames[gameID])
			}
		}
		sb.WriteString(fmt.Sprintf("%d. Usuario: %d (Similitud: %.2f) | Juegos en común: %s\n", i+1, simUser.UserID, simUser.Similarity, strings.Join(commonGames, ", ")))
	}

	// Escribe los resultados de todas las pruebas de rendimiento
	sb.WriteString("\n--- Rendimiento ---\n")
	sb.WriteString(fmt.Sprintf("Tiempo de ejecucion secuencial (base): %v\n", durationSequential))

	// Para un orden consistente, ordenamos las llaves del mapa (2, 4, 8)
	counts := make([]int, 0, len(concurrentResults))
	for k := range concurrentResults {
		counts = append(counts, k)
	}
	sort.Ints(counts)

	for _, count := range counts {
		duration := concurrentResults[count]
		speedup := float64(durationSequential) / float64(duration)
		sb.WriteString(fmt.Sprintf("Tiempo concurrente (%d goroutines): %v (Speedup: %.2fx)\n", count, duration, speedup))
	}

	dirPath := "Pruebas"
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		fmt.Printf("\nERROR al crear el directorio: %v\n", err)
		return
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("Prueba_%s.txt", timestamp)
	filePath := filepath.Join(dirPath, filename)
	if err := os.WriteFile(filePath, []byte(sb.String()), 0644); err != nil {
		fmt.Printf("\nERROR al guardar el archivo: %v\n", err)
		return
	}
	fmt.Printf("\n- Resultados del benchmark guardados en: %s\n", filePath)
}

// ============= FUNCIÓN PRINCIPAL =============
func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("========================================")
	fmt.Println("  SISTEMA DE RECOMENDACIÓN STEAM")
	fmt.Println("========================================\n")

	reviews := loadAllReviews(CsvFilePath, MaxReviewsToLoad)
	if reviews == nil {
		fmt.Printf("ERROR: No se pudieron cargar las reseñas.\n")
		return
	}
	reviewsCount := len(reviews)
	fmt.Printf("- Cargadas %d reseñas válidas\n\n", len(reviews))

	fmt.Println("2. Construyendo matriz usuario-juego...")
	matrix := buildRatingMatrix(reviews, MinGamesPerUser)
	reviews = nil
	runtime.GC()
	fmt.Printf("- Usuarios en la matriz: %d\n", len(matrix.UserList))
	fmt.Printf("- Juegos en la matriz: %d\n\n", len(matrix.GameList))

	fmt.Println("3. Normalizando ratings...")
	normalized := normalizeMatrix(matrix)
	fmt.Println("- Normalización completada\n")

	if len(matrix.UserList) == 0 {
		fmt.Println("ERROR: No hay usuarios que cumplan el criterio mínimo.")
		return
	}

	var targetUser int64
	if TargetUserID > 0 {
		fmt.Printf("4. Buscando usuario específico (ID %d)...\n", TargetUserID)
		if _, ok := matrix.Data[TargetUserID]; ok {
			targetUser = TargetUserID
		} else {
			fmt.Printf("ERROR: El usuario con ID %d no fue encontrado. Se usará uno aleatorio.\n", TargetUserID)
			targetUser = matrix.UserList[rand.Intn(len(matrix.UserList))]
		}
	} else {
		fmt.Println("4. Seleccionando usuario aleatorio...")
		targetUser = matrix.UserList[rand.Intn(len(matrix.UserList))]
	}

	userRatings := matrix.Data[targetUser]
	avgRating := 0.0
	for _, rating := range userRatings {
		avgRating += rating
	}
	avgRating /= float64(len(userRatings))
	fmt.Printf("- Usuario seleccionado: %d\n", targetUser)
	fmt.Printf("- Juegos jugados: %d\n", len(userRatings))
	fmt.Printf("- Rating promedio: %.2f\n\n", avgRating)

	userGames := make([]GameScore, 0, len(userRatings))
	for gameID, rating := range userRatings {
		userGames = append(userGames, GameScore{Name: matrix.GameNames[gameID], Rating: rating})
	}
	sort.Slice(userGames, func(i, j int) bool { return userGames[i].Rating > userGames[j].Rating })
	fmt.Println("Top 5 juegos del usuario (rating 1-10):")
	for i := 0; i < 5 && i < len(userGames); i++ {
		fmt.Printf("  %d. %s (Rating: %.2f)\n", i+1, userGames[i].Name, userGames[i].Rating)
	}
	fmt.Println()

	// --- BENCHMARK DE RENDIMIENTO ---
	fmt.Println("5. USUARIOS SIMILARES - BENCHMARK DE RENDIMIENTO")

	// --- Ejecución Secuencial ---
	fmt.Println("\n- Ejecutando búsqueda secuencial ")
	startSequential := time.Now()
	findSimilarUsersSequential(matrix, normalized, targetUser)
	durationSequential := time.Since(startSequential)
	fmt.Printf("- Búsqueda secuencial completada en %v.\n", durationSequential)

	// --- Bucle de Pruebas Concurrentes ---
	goroutineCountsToTest := []int{2, 4, 8}
	concurrentResults := make(map[int]time.Duration)
	var similarUsers []UserSimilarity

	for _, count := range goroutineCountsToTest {
		fmt.Printf("\n- Ejecutando búsqueda CONCURRENTE con %d goroutines...\n", count)
		startConcurrent := time.Now()
		similarUsers = findSimilarUsersConcurrent(matrix, normalized, targetUser, count)
		durationConcurrent := time.Since(startConcurrent)
		concurrentResults[count] = durationConcurrent

		if durationSequential > 0 {
			speedup := float64(durationSequential) / float64(durationConcurrent)
			fmt.Printf("- Completado en %v (Speedup: %.2fx).\n", durationConcurrent, speedup)
		} else {
			fmt.Printf("- Completado en %v.\n", durationConcurrent)
		}
	}
	fmt.Println()

	fmt.Println("6. Generando recomendaciones...")
	recommendations := generateRecommendations(matrix, targetUser, similarUsers)
	if len(recommendations) == 0 {
		fmt.Println("No se pudieron generar recomendaciones. Intenta con otro usuario.")
	} else {
		fmt.Printf("- Generadas %d recomendaciones.\n\n", len(recommendations))
		fmt.Println("========================================")
		fmt.Printf("        TOP %d RECOMENDACIONES\n", NumRecommendations)
		fmt.Println("========================================\n")
		for i := 0; i < NumRecommendations && i < len(recommendations); i++ {
			rec := recommendations[i]
			fmt.Printf("%d. %s\n", i+1, rec.GameName)
			fmt.Printf("   Score Predicho: %.2f | Basado en %d usuarios similares\n\n", rec.PredScore, rec.NumVoters)
		}
	}

	fmt.Println("========================================")
	fmt.Println("       TOP 5 USUARIOS MÁS SIMILARES")
	fmt.Println("========================================\n")
	if len(similarUsers) > 0 {
		numTopUsersToShow := 5
		if len(similarUsers) < numTopUsersToShow {
			numTopUsersToShow = len(similarUsers)
		}
		for i := 0; i < numTopUsersToShow; i++ {
			simUser := similarUsers[i]
			simUserData := matrix.Data[simUser.UserID]
			fmt.Printf("%d. Usuario: %d (Similitud: %.2f)\n", i+1, simUser.UserID, simUser.Similarity)
			commonGames := make([]string, 0)
			for gameID := range userRatings {
				if _, exists := simUserData[gameID]; exists {
					commonGames = append(commonGames, matrix.GameNames[gameID])
				}
			}
			if len(commonGames) > 0 {
				fmt.Printf("   Juegos en común: %s\n\n", strings.Join(commonGames, ", "))
			} else {
				fmt.Println()
			}
		}
	} else {
		fmt.Println("No se encontraron usuarios suficientemente similares para analizar.")
	}

	// Guardar resultados en archivo
	saveResultsToFile(
		reviewsCount,
		matrix,
		targetUser,
		userGames,
		avgRating,
		recommendations,
		similarUsers,
		durationSequential,
		concurrentResults,
	)

	fmt.Println("========================================")
}

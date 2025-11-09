package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Constante para manejar la cantidad máxima de registros a procesar del archivo CSV (opcional)
// Si se pone en -1, se procesan todos los registros
const MaxRecords = 100

// Definimos la estructura para almacenar los ratings
type Rating struct {
	UserID  int
	MovieID int
	Rating  float32
}

func preprocesamiento(path string) []Rating {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error al abrir el archivo: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read() // Saltamos el primer registro, que es el encabezado

	// Estructuras para el seguimiento y estadísticas
	seen := make(map[string]bool)
	userSet := make(map[int]bool)
	movieSet := make(map[int]bool)

	// Slice para almacenar datos limpios
	var cleanData []Rating

	// Contadores para métricas
	total, invalid, duplicates := 0, 0, 0

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		if MaxRecords != -1 && total >= MaxRecords {
			break
		}

		total++

		// Verificamos la longitud del registro y la presencia de valores vacíos
		if len(record) < 3 || record[0] == "" || record[1] == "" || record[2] == "" {
			invalid++
			continue
		}

		// Realizamos la conversión de tipos de datos
		userID, err1 := strconv.Atoi(record[0])
		movieID, err2 := strconv.Atoi(record[1])
		rating64, err3 := strconv.ParseFloat(record[2], 32)

		// Verificamos errores de conversión
		if err1 != nil || err2 != nil || err3 != nil {
			invalid++
			continue
		}

		// Convertimos rating a float32
		rating := float32(rating64)

		// Validamos el rango de rating
		if rating < 0.5 || rating > 5.0 {
			invalid++
			continue
		}

		// Detectamos duplicados, e ignoramos si ya existe
		key := fmt.Sprintf("%d-%d", userID, movieID)
		if seen[key] {
			duplicates++
			continue
		}
		seen[key] = true

		// Agregamos el registro limpio
		cleanData = append(cleanData, Rating{userID, movieID, rating})

		// Actualizamos conjuntos de usuarios y películas únicos
		userSet[userID] = true
		movieSet[movieID] = true
	}

	// Mostramos estadísticas del preprocesamiento
	fmt.Println("Resultados del Preprocesamiento:")
	fmt.Printf("\nTotal de registros leídos: %d\n", total)
	fmt.Printf("Registros válidos: %d\n", len(cleanData))
	fmt.Printf("Registros inválidos: %d (%.2f%%)\n", invalid, float64(invalid)/float64(total)*100)
	fmt.Printf("Registros duplicados: %d (%.2f%%)\n", duplicates, float64(duplicates)/float64(total)*100)
	fmt.Printf("Usuarios únicos: %d\n", len(userSet))
	fmt.Printf("Películas únicas: %d\n", len(movieSet))

	return cleanData
}

func generarMatriz(data []Rating) ([][]float32, []int, []int) {
	userIndex := make(map[int]int)
	movieIndex := make(map[int]int)
	var users []int
	var movies []int

	// Asignamos índices secuenciales a usuarios y películas, mediante mapas
	for _, r := range data {
		if _, exists := userIndex[r.UserID]; !exists {
			userIndex[r.UserID] = len(users)
			users = append(users, r.UserID)
		}
		if _, exists := movieIndex[r.MovieID]; !exists {
			movieIndex[r.MovieID] = len(movies)
			movies = append(movies, r.MovieID)
		}
	}

	// Creamos una matriz de ceros
	matriz := make([][]float32, len(users))
	for i := range matriz {
		matriz[i] = make([]float32, len(movies))
	}

	// Llenamos la matriz con los ratings
	for _, r := range data {
		u := userIndex[r.UserID]
		m := movieIndex[r.MovieID]
		matriz[u][m] = r.Rating
	}

	// Normalización Min–Max [0,1]
	const min, max = float32(0.5), float32(5.0)
	for i := range matriz {
		for j := range matriz[i] {
			val := matriz[i][j]
			if val == 0 {
				continue // Para mantener los ceros originales
			}
			matriz[i][j] = (val - min) / (max - min)
		}
	}

	fmt.Printf("\nMatriz generada: %d usuarios x %d películas\n", len(users), len(movies))
	return matriz, users, movies
}

func guardarMatrizCSV(matriz [][]float32, users []int, movies []int, outputPath string) {
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Error al crear archivo CSV: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// Encabezado: columnas por índices de películas
	header := []string{"userIndex"}
	for j := range movies {
		header = append(header, strconv.Itoa(j))
	}
	writer.Write(header)

	// Filas: índice de usuario + ratings
	for i := range users {
		row := []string{strconv.Itoa(i)}
		for _, rating := range matriz[i] {
			row = append(row, fmt.Sprintf("%.3f", rating))
		}
		writer.Write(row)
	}

	writer.Flush()
	fmt.Printf("Archivo matriz guardado: %s\n", outputPath)
}

func guardarMapeos(users []int, movies []int) {
	// Mapeo de usuarios
	uFile, err := os.Create("usuarios_mapping.csv")
	if err != nil {
		log.Fatalf("Error al crear usuarios_mapping.csv: %v", err)
	}
	defer uFile.Close()

	uWriter := csv.NewWriter(uFile)
	uWriter.Write([]string{"userIndex", "userId"})
	for i, u := range users {
		uWriter.Write([]string{strconv.Itoa(i), strconv.Itoa(u)})
	}
	uWriter.Flush()

	// Mapeo de películas
	mFile, err := os.Create("peliculas_mapping.csv")
	if err != nil {
		log.Fatalf("Error al crear peliculas_mapping.csv: %v", err)
	}
	defer mFile.Close()

	mWriter := csv.NewWriter(mFile)
	mWriter.Write([]string{"movieIndex", "movieId"})
	for j, m := range movies {
		mWriter.Write([]string{strconv.Itoa(j), strconv.Itoa(m)})
	}
	mWriter.Flush()

	fmt.Println("Mapeos guardados: usuarios_mapping.csv y peliculas_mapping.csv")
}

func main() {
	// Función de preprocesamiento
	cleanData := preprocesamiento("ratings.csv")

	fmt.Printf("Se obtuvieron %d registros limpios.\n", len(cleanData))

	// Generación de la matriz de usuarios y películas
	matriz, users, movies := generarMatriz(cleanData)

	// Guardar la matriz en un archivo CSV
	guardarMatrizCSV(matriz, users, movies, "matriz_usuarios_peliculas.csv")

	// Guardar los mapeos de usuarios y películas
	guardarMapeos(users, movies)
}

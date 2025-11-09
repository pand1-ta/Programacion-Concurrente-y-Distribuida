package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

// CARGA DE MATRIZ
func cargarMatriz(path string) [][]float32 {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error al abrir matriz: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	registros, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error al leer CSV: %v", err)
	}

	var matriz [][]float32
	for i, fila := range registros {
		if i == 0 { // saltar encabezado
			continue
		}
		var filaNum []float32
		for j := 1; j < len(fila); j++ {
			val, _ := strconv.ParseFloat(fila[j], 32)
			filaNum = append(filaNum, float32(val))
		}
		matriz = append(matriz, filaNum)
	}
	fmt.Printf("Matriz cargada: %d usuarios x %d películas\n", len(matriz), len(matriz[0]))
	return matriz
}

// CÁLCULO DE SIMILITUD COSENO
func similitudCoseno(u, v []float32) float32 {
	var dot, normU, normV float32
	for i := 0; i < len(u); i++ {
		dot += u[i] * v[i]
		normU += u[i] * u[i]
		normV += v[i] * v[i]
	}
	if normU == 0 || normV == 0 {
		return 0
	}
	return dot / (float32(math.Sqrt(float64(normU))) * float32(math.Sqrt(float64(normV))))
}

// CÁLCULO SECUENCIAL
func calcularSimilitudSecuencial(matriz [][]float32) [][]float32 {
	n := len(matriz)
	sim := make([][]float32, n)
	for i := range sim {
		sim[i] = make([]float32, n)
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			simVal := similitudCoseno(matriz[i], matriz[j])
			sim[i][j] = simVal
			sim[j][i] = simVal
		}
	}
	return sim
}

// ESTRUCTURA PARA COMUNICAR RESULTADOS
type resultado struct {
	i, j  int
	valor float32
}

// PROCESAMIENTO CONCURRENTE USANDO CHANNELS
func procesarSubmatriz(id int, matriz [][]float32, inicio, fin int, ch chan<- resultado) {
	for i := inicio; i < fin; i++ {
		for j := i + 1; j < len(matriz); j++ {
			val := similitudCoseno(matriz[i], matriz[j])
			ch <- resultado{i, j, val}
		}
	}
	fmt.Printf("Goroutine %d procesó filas [%d:%d)\n", id, inicio, fin)
}

func calcularSimilitudConcurrente(matriz [][]float32, numWorkers int) [][]float32 {
	n := len(matriz)
	sim := make([][]float32, n)
	for i := range sim {
		sim[i] = make([]float32, n)
	}

	ch := make(chan resultado)
	chunk := n / numWorkers

	// Lanzar workers
	for i := 0; i < numWorkers; i++ {
		inicio := i * chunk
		fin := inicio + chunk
		if i == numWorkers-1 {
			fin = n
		}
		go procesarSubmatriz(i, matriz, inicio, fin, ch)
	}

	// Número total de pares (i,j)
	total := n * (n - 1) / 2
	for k := 0; k < total; k++ {
		r := <-ch
		sim[r.i][r.j] = r.valor
		sim[r.j][r.i] = r.valor
	}

	close(ch)
	return sim
}

// MEDICIÓN DE SPEEDUP Y ESCALABILIDAD
func medirTiempos(matriz [][]float32) {
	fmt.Println("\nMEDICIÓN DE SPEEDUP CON SIMILITUD COSENO")

	// Ejecución secuencial
	inicioSeq := time.Now()
	_ = calcularSimilitudSecuencial(matriz)
	tiempoSeq := time.Since(inicioSeq)
	fmt.Printf("Tiempo SECUENCIAL: %v\n", tiempoSeq)

	// Ejecución concurrente
	for _, workers := range []int{2, 4, 8} {
		inicioPar := time.Now()
		_ = calcularSimilitudConcurrente(matriz, workers)
		tiempoPar := time.Since(inicioPar)
		speedup := float64(tiempoSeq) / float64(tiempoPar)

		fmt.Printf("\nGoroutines: %d\n", workers)
		fmt.Printf("Tiempo paralelo: %v\n", tiempoPar)
		fmt.Printf("Speedup: %.2fx\n", speedup)
	}
}

// MAIN
func main() {
	matriz := cargarMatriz("matriz_usuarios_peliculas.csv")
	medirTiempos(matriz)
}

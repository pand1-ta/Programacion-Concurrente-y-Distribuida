package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type PartialResult struct {
	dotProduct float64
	normA      float64
	normB      float64
}

// Cálculo secuencial
func cosineSimilaritySequential(a, b []float64) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Cálculo concurrente
func cosineSimilarityConcurrent(a, b []float64, numGoroutines int) float64 {
	chunkSize := len(a) / numGoroutines
	results := make(chan PartialResult, numGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numGoroutines-1 {
			end = len(a)
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			var pr PartialResult
			for j := start; j < end; j++ {
				pr.dotProduct += a[j] * b[j]
				pr.normA += a[j] * a[j]
				pr.normB += b[j] * b[j]
			}
			results <- pr
		}(start, end)
	}

	wg.Wait()
	close(results)

	var totalDot, totalNormA, totalNormB float64
	for r := range results {
		totalDot += r.dotProduct
		totalNormA += r.normA
		totalNormB += r.normB
	}

	return totalDot / (math.Sqrt(totalNormA) * math.Sqrt(totalNormB))
}

func main() {
	vectorSizes := []int{1_000_000, 2_500_000, 3_000_000, 4_000_000, 5_000_000, 10_000_000, 20_000_000}
	goroutineCounts := []int{2, 4, 8}
	const seed = 42
	const maxVal = 1000

	for _, size := range vectorSizes {
		// Preparar vectores aleatorios
		rand.Seed(seed)
		a := make([]float64, size)
		b := make([]float64, size)
		for i := 0; i < size; i++ {
			a[i] = float64(rand.Intn(maxVal + 1))
			b[i] = float64(rand.Intn(maxVal + 1))
		}

		// Calcular secuencial
		startSeq := time.Now()
		simSeq := cosineSimilaritySequential(a, b)
		durationSeq := time.Since(startSeq)

		for _, g := range goroutineCounts {
			startConc := time.Now()
			simConc := cosineSimilarityConcurrent(a, b, g)
			durationConc := time.Since(startConc)

			var speedup string
			if durationConc.Milliseconds() == 0 {
				speedup = "+Inf"
			} else {
				speedupVal := float64(durationSeq.Microseconds()) / float64(durationConc.Microseconds())
				speedup = fmt.Sprintf("%.2fx", speedupVal)
			}

			fmt.Printf("\nVectorSize: %d | Goroutines: %d\n", size, g)
			fmt.Printf("→ Similitud Secuencial:  %.6f | Tiempo: %v\n", simSeq, durationSeq)
			fmt.Printf("→ Similitud Concurrente: %.6f | Tiempo: %v\n", simConc, durationConc)
			fmt.Printf("→ Speedup: %s\n", speedup)
		}
	}
}

package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Estructura para guardar resultados parciales calculados por cada goroutine
type partial struct {
	sumX, sumY, sumXY, sumX2, sumY2 float64
	n                               int
}

// Función worker que calcula los valores parciales para un segmento del vector
func worker(start, end int, x, y []float64, results chan<- partial, wg *sync.WaitGroup) {
	defer wg.Done()
	var p partial
	for i := start; i < end; i++ {
		xi, yi := x[i], y[i]
		p.sumX += xi
		p.sumY += yi
		p.sumXY += xi * yi
		p.sumX2 += xi * xi
		p.sumY2 += yi * yi
		p.n++
	}
	results <- p
}

// Cálculo secuencial del coeficiente de Pearson
func pearsonSecuencial(x, y []float64) float64 {
	n := float64(len(x))
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	// Acumulación de sumatorias necesarias
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	// Fórmula de Pearson
	num := n*sumXY - sumX*sumY
	den := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	// Verificamos que no haya división por cero
	if den == 0 {
		return 0
	}
	return num / den
}

// Cálculo concurrente del coeficiente de Pearson con múltiples goroutines
func pearsonConcurrente(x, y []float64, numWorkers int) float64 {
	n := len(x)
	var wg sync.WaitGroup
	results := make(chan partial, numWorkers)
	chunk := n / numWorkers // División del trabajo por goroutine

	// Lanzamiento de workers para cada segmento del vector
	for w := 0; w < numWorkers; w++ {
		start := w * chunk
		end := start + chunk
		if w == numWorkers-1 {
			end = n // Última goroutine toma el resto
		}
		wg.Add(1)
		go worker(start, end, x, y, results, &wg)
	}

	wg.Wait()      // Esperamos a que terminen todas
	close(results) // Cerramos el canal

	// Acumulación de resultados parciales
	var total partial
	for r := range results {
		total.sumX += r.sumX
		total.sumY += r.sumY
		total.sumXY += r.sumXY
		total.sumX2 += r.sumX2
		total.sumY2 += r.sumY2
		total.n += r.n
	}

	// Cálculo final con la fórmula de Pearson
	N := float64(total.n)
	num := N*total.sumXY - total.sumX*total.sumY
	den := math.Sqrt((N*total.sumX2 - total.sumX*total.sumX) * (N*total.sumY2 - total.sumY*total.sumY))
	if den == 0 {
		return 0
	}
	return num / den
}

// Función principal que ejecuta pruebas con distintos tamaños de vectores y niveles de concurrencia
func main() {
	// Distintos tamaños de vectores a probar
	vectorSizes := []int{1_000_000, 2_500_000, 3_000_000, 4_000_000, 5_000_000, 10_000_000, 20_000_000}
	// Cantidad de goroutines a probar en concurrencia
	goroutineCounts := []int{2, 4, 8}
	const seed = 42
	const maxVal = 1000

	for _, size := range vectorSizes {
		// Generación de datos aleatorios reproducibles con semilla fija
		rand.Seed(seed)
		x := make([]float64, size)
		y := make([]float64, size)
		for i := 0; i < size; i++ {
			x[i] = float64(rand.Intn(maxVal + 1))
			y[i] = float64(rand.Intn(maxVal + 1))
		}

		// Pearson secuencial
		startSeq := time.Now()
		resultSeq := pearsonSecuencial(x, y)
		durationSeq := time.Since(startSeq)

		// Pearson concurrente con diferentes números de goroutines
		for _, g := range goroutineCounts {
			startConc := time.Now()
			resultConc := pearsonConcurrente(x, y, g)
			durationConc := time.Since(startConc)

			// Cálculo del speedup (con protección por si el tiempo concurrente es muy bajo)
			var speedup string
			if durationConc.Microseconds() == 0 {
				speedup = "+Inf"
			} else {
				speedupVal := float64(durationSeq.Microseconds()) / float64(durationConc.Microseconds())
				speedup = fmt.Sprintf("%.2fx", speedupVal)
			}

			// Mostrar resultados
			fmt.Printf("\nVectorSize: %d | Goroutines: %d\n", size, g)
			fmt.Printf("→ Pearson Secuencial:  %.6f | Tiempo: %v\n", resultSeq, durationSeq)
			fmt.Printf("→ Pearson Concurrente: %.6f | Tiempo: %v\n", resultConc, durationConc)
			fmt.Printf("→ Speedup: %s\n", speedup)
		}
	}
}

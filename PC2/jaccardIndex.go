package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Función para calcular el índice de Jaccard de forma secuencial
func jaccardSecuencial(a, b []int) float64 {
	// Convertimos los slices en sets usando mapas
	setA := make(map[int]bool)
	setB := make(map[int]bool)

	for _, v := range a {
		setA[v] = true
	}
	for _, v := range b {
		setB[v] = true
	}

	// Calculamos la intersección y la unión
	intersection := 0
	union := make(map[int]bool)

	for v := range setA {
		union[v] = true
		if setB[v] {
			intersection++
		}
	}
	for v := range setB {
		union[v] = true
	}

	// Índice de Jaccard = intersección / unión
	return float64(intersection) / float64(len(union))
}

// Función para calcular el índice de Jaccard de forma concurrente usando worker pool
func jaccardConcurrente(a, b []int, numWorkers int) float64 {
	// Función para construir un conjunto único desde un slice, de forma concurrente
	makeSet := func(arr []int, numWorkers int) map[int]bool {
		set := make(map[int]bool)
		var wg sync.WaitGroup
		ch := make(chan int, len(arr))

		// Dividimos el array en chunks para paralelizar
		chunkSize := (len(arr) + numWorkers - 1) / numWorkers
		for i := 0; i < len(arr); i += chunkSize {
			end := i + chunkSize
			if end > len(arr) {
				end = len(arr)
			}

			wg.Add(1)
			go func(chunk []int) {
				defer wg.Done()
				seen := make(map[int]bool)
				for _, v := range chunk {
					if !seen[v] {
						seen[v] = true
						ch <- v
					}
				}
			}(arr[i:end])
		}

		// Cerramos el canal cuando terminan los workers
		go func() {
			wg.Wait()
			close(ch)
		}()

		// Recolectamos los elementos únicos
		for v := range ch {
			set[v] = true
		}
		return set
	}

	// Creamos los sets de forma concurrente
	setA := makeSet(a, numWorkers)
	setB := makeSet(b, numWorkers)

	// Tipo para guardar resultados parciales
	type partial struct {
		inter int
		union map[int]bool
	}

	ch := make(chan partial, numWorkers)
	var wg sync.WaitGroup

	// Convertimos el setA en slice para dividirlo
	elems := make([]int, 0, len(setA))
	for v := range setA {
		elems = append(elems, v)
	}

	// Dividimos los datos de setA entre los workers
	chunkSize := (len(elems) + numWorkers - 1) / numWorkers
	for i := 0; i < len(elems); i += chunkSize {
		end := i + chunkSize
		if end > len(elems) {
			end = len(elems)
		}
		wg.Add(1)
		go func(chunk []int) {
			defer wg.Done()
			localUnion := make(map[int]bool)
			inter := 0
			for _, v := range chunk {
				localUnion[v] = true
				if setB[v] {
					inter++
				}
			}
			ch <- partial{inter, localUnion}
		}(elems[i:end])
	}

	// Agregamos todos los elementos de setB a la unión (en otra goroutine)
	wg.Add(1)
	go func() {
		defer wg.Done()
		localUnion := make(map[int]bool)
		for v := range setB {
			localUnion[v] = true
		}
		ch <- partial{0, localUnion}
	}()

	// Cerramos el canal cuando todas las goroutines terminan
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Reducimos los resultados parciales
	intersection := 0
	union := make(map[int]bool)
	for p := range ch {
		intersection += p.inter
		for v := range p.union {
			union[v] = true
		}
	}

	// Cálculo final del índice de Jaccard
	return float64(intersection) / float64(len(union))
}

// Función principal que realiza las pruebas con distintos tamaños y niveles de concurrencia
func main() {
	// Distintos tamaños de vectores a probar
	vectorSizes := []int{1_000_000, 2_500_000, 3_000_000, 4_000_000, 5_000_000, 10_000_000, 20_000_000}
	goroutineCounts := []int{2, 4, 8}
	const seed = 42
	const maxVal = 1_000_000

	for _, size := range vectorSizes {
		// Generamos vectores aleatorios con la misma semilla para consistencia
		rand.Seed(seed)
		a := make([]int, size)
		b := make([]int, size)
		for i := 0; i < size; i++ {
			a[i] = rand.Intn(maxVal)
			b[i] = rand.Intn(maxVal)
		}

		// Medimos tiempo para la versión secuencial
		startSeq := time.Now()
		jaccSeq := jaccardSecuencial(a, b)
		durationSeq := time.Since(startSeq)

		// Ejecutamos versión concurrente con distintos niveles de goroutines
		for _, g := range goroutineCounts {
			startConc := time.Now()
			jaccConc := jaccardConcurrente(a, b, g)
			durationConc := time.Since(startConc)

			// Cálculo del speedup
			var speedup string
			if durationConc.Microseconds() == 0 {
				speedup = "+Inf"
			} else {
				val := float64(durationSeq.Microseconds()) / float64(durationConc.Microseconds())
				speedup = fmt.Sprintf("%.2fx", val)
			}

			// Imprimimos los resultados
			fmt.Printf("\nVectorSize: %d | Goroutines: %d\n", size, g)
			fmt.Printf("→ Jaccard Secuencial:  %.6f | Tiempo: %v\n", jaccSeq, durationSeq)
			fmt.Printf("→ Jaccard Concurrente: %.6f | Tiempo: %v\n", jaccConc, durationConc)
			fmt.Printf("→ Speedup: %s\n", speedup)
		}
	}
}

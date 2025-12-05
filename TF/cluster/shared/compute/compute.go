package compute

import (
	"math"
	"sort"
)

// --------------------------------------------
// UTILIDAD: similitud coseno entre dos vectores
// --------------------------------------------
func cosine(u, v []float64) float64 {
	var dot, nu, nv float64
	for i := range u {
		dot += u[i] * v[i]
		nu += u[i] * u[i]
		nv += v[i] * v[i]
	}

	if nu == 0 || nv == 0 {
		return 0
	}

	return dot / (math.Sqrt(nu) * math.Sqrt(nv))
}

// ---------------------------------------------------
// SIMILITUD PARA UN USUARIO CONTRA TODOS LOS DEMÁS
// ---------------------------------------------------
func CosineSimilarityForUser(matrix [][]float64, userIndex int) []float64 {
	n := len(matrix)
	sims := make([]float64, n)

	target := matrix[userIndex]

	for i := 0; i < n; i++ {
		if i == userIndex {
			sims[i] = -1 // para evitar que sea elegido como su propio vecino
			continue
		}
		sims[i] = cosine(target, matrix[i])
	}

	return sims
}

// ---------------------------------------------------
// MATRIZ COMPLETA DE SIMILITUD
// (solo si la API lo necesita)
// ---------------------------------------------------
func CosineSimilarityMatrix(matrix [][]float64) []float64 {
	n := len(matrix)
	result := make([]float64, 0, n*n)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			result = append(result, cosine(matrix[i], matrix[j]))
		}
	}

	return result
}

// ---------------------------------------------------
// PREDICCIÓN BASADA EN K VECINOS
// ---------------------------------------------------
func PredictRatings(matrix [][]float64, sims []float64, userIndex, k int) []float64 {
	// 1. Obtener índices de vecinos ordenados
	idxs := sortIndexesDescending(sims)

	// Excluir el propio usuario de la lista de vecinos
	neighbors := make([]int, 0, len(idxs))
	for _, i := range idxs {
		if i == userIndex {
			continue
		}
		neighbors = append(neighbors, i)
	}

	// Ajustar k si es mayor que el número de vecinos disponibles
	if k > len(neighbors) {
		k = len(neighbors)
	}

	// 2. Tomar los K mejores vecinos (puede ser 0)
	best := neighbors[:k]

	target := matrix[userIndex]
	m := len(target)
	preds := make([]float64, m)

	for movie := 0; movie < m; movie++ {

		// si ya tiene valor, mantenemos su rating
		if target[movie] > 0 {
			preds[movie] = target[movie]
			continue
		}

		var num, den float64

		// ponderar por similitud
		for _, neighbor := range best {
			rating := matrix[neighbor][movie]
			if rating == 0 {
				continue
			}

			num += sims[neighbor] * rating
			den += math.Abs(sims[neighbor])
		}

		if den == 0 {
			preds[movie] = 0
		} else {
			preds[movie] = num / den
		}
	}

	return preds
}

// ---------------------------------------------------
// ORDENAR PELÍCULAS POR PUNTAJE
// ---------------------------------------------------
func SortIndexesByScore(scores []float64) []int {
	return sortIndexesDescending(scores)
}

// Utilidad: ordenar de mayor a menor
func sortIndexesDescending(values []float64) []int {
	idx := make([]int, len(values))
	for i := range idx {
		idx[i] = i
	}

	sort.Slice(idx, func(i, j int) bool {
		return values[idx[i]] > values[idx[j]]
	})

	return idx
}

package dispatcher

import (
	"fmt"
	"sdr/cluster/coordinator/internal/compute"
	"sdr/cluster/coordinator/internal/models"
)

// Process es el punto de entrada del coordinador para procesar solicitudes
func Process(msg models.TaskMessage) (models.CoordinatorResponse, error) {
	switch msg.Type {
	case models.RequestSimilarity:
		return processSimilarity(msg)
	case models.RequestRecommendation:
		return processRecommendation(msg)
	default:
		return models.CoordinatorResponse{}, fmt.Errorf("tipo de solicitud no reconocido: %s", msg.Type)
	}
}

// -------------------------------------------
// PROCESAR SIMILITUD
// -------------------------------------------
func processSimilarity(msg models.TaskMessage) (models.CoordinatorResponse, error) {

	// Por ahora no dividimos: fallback local
	result := compute.CosineSimilarityMatrix(msg.Matrix)

	return models.CoordinatorResponse{
		Result: result,
	}, nil
}

// -------------------------------------------
// PROCESAR RECOMENDACIÓN
// -------------------------------------------
func processRecommendation(msg models.TaskMessage) (models.CoordinatorResponse, error) {

	// Paso 1: obtener similitudes
	sims := compute.CosineSimilarityForUser(msg.Matrix, msg.UserIndex)

	// Paso 2: obtener predicciones ponderadas usando K vecinos
	preds := compute.PredictRatings(msg.Matrix, sims, msg.UserIndex, msg.K)

	// Paso 3: ordenar películas por puntaje
	indexes := compute.SortIndexesByScore(preds)

	return models.CoordinatorResponse{
		Result:  preds,
		Indexes: indexes,
	}, nil
}

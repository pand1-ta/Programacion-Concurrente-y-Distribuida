package dispatcher

import (
	"fmt"
	"log"
	"sync"

	"sdr/cluster/coordinator/internal/tcpclient"
	"sdr/cluster/shared/compute"
	"sdr/cluster/shared/models"
)

// Lista de direcciones de los Workers en la red Docker
var workerAddrs = []string{
	"sdr_worker1:9000",
	"sdr_worker2:9000",
	"sdr_worker3:9000",
	"sdr_worker4:9000",
	"sdr_worker5:9000",
	"sdr_worker6:9000",
	"sdr_worker7:9000",
	"sdr_worker8:9000",
}

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
// PROCESAR SIMILITUD (local o distribuido)
// -------------------------------------------
func processSimilarity(msg models.TaskMessage) (models.CoordinatorResponse, error) {
	// Por simplicidad, lo mantenemos local
	result := compute.CosineSimilarityMatrix(msg.Matrix)

	return models.CoordinatorResponse{
		Result: result,
	}, nil
}

// -------------------------------------------
// PROCESAR RECOMENDACIÓN (distribuido)
// -------------------------------------------
func processRecommendation(msg models.TaskMessage) (models.CoordinatorResponse, error) {
	log.Println("Iniciando processRecommendation...")
	var wg sync.WaitGroup
	results := make(chan []float64, len(workerAddrs))

	// Enviar la tarea a los 8 Workers en paralelo
	for _, addr := range workerAddrs {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			log.Printf("Intentando conectar con worker %s...\n", a)
			resp, err := tcpclient.SendTask(a, msg)
			if err != nil {
				log.Printf("Error comunicando con %s: %v\n", a, err)
				return
			}
			log.Printf("Respuesta recibida de %s con %d resultados\n", a, len(resp.Result))
			if len(resp.Result) > 0 {
				results <- resp.Result
			}
		}(addr)
	}

	wg.Wait()
	close(results)
	log.Println("Todos los goroutines completados, combinando resultados...")

	// Combinar resultados parciales (promedio)
	var combined []float64
	count := 0
	for r := range results {
		if combined == nil {
			combined = make([]float64, len(r))
		}
		for i := range r {
			combined[i] += r[i]
		}
		count++
	}

	if count == 0 {
		log.Println("No se recibieron resultados de ningún worker.")
		return models.CoordinatorResponse{}, fmt.Errorf("no se recibieron resultados de los workers")
	}

	for i := range combined {
		combined[i] /= float64(count)
	}

	indexes := compute.SortIndexesByScore(combined)
	log.Println("Recomendaciones combinadas, enviando respuesta a la API...")

	return models.CoordinatorResponse{
		Result:  combined,
		Indexes: indexes,
	}, nil
}

package models

// Tipo de operación que la API pide al coordinador.
type RequestType string

const (
	RequestSimilarity     RequestType = "SIMILARITY"
	RequestRecommendation RequestType = "RECOMMENDATION"
)

// Mensaje base que la API envía al coordinador vía TCP
type TaskMessage struct {
	Type      RequestType `json:"type"`
	Matrix    [][]float64 `json:"matrix"`    // matriz completa usuario–película
	UserIndex int         `json:"userIndex"` // solo para recomendación
	K         int         `json:"k"`         // vecinos
}

// --- Chunking ---

type Chunk struct {
	ID        int         `json:"id"`
	Start     int         `json:"start"`
	End       int         `json:"end"`
	Matrix    [][]float64 `json:"matrix"`
	UserIndex int         `json:"userIndex"` // solo se usa para recomendación
	K         int         `json:"k"`
}

// --- Worker: mensaje enviado por el coordinador ---

type WorkerTask struct {
	Chunk Chunk `json:"chunk"`
}

// --- Worker: resultado enviado al coordinador ---

type WorkerResult struct {
	ChunkID int       `json:"chunkId"`
	Values  []float64 `json:"values"` // similitud o predicciones
}

// --- Respuesta final para la API ---

type CoordinatorResponse struct {
	Result  []float64 `json:"result"`
	Indexes []int     `json:"indexes,omitempty"` // para recomendación (top-N ordenado)
}

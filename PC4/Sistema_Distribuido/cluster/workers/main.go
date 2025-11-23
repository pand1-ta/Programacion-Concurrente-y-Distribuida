package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"sdr/cluster/shared/compute"
	"sdr/cluster/shared/models"
)

func main() {
	port := "9000"
	if p := os.Getenv("WORKER_PORT"); p != "" {
		port = p
	}

	fmt.Printf("Worker escuchando en puerto %s...\n", port)
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Printf("Error iniciando el listener: %v\n", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Error aceptando conexión: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Leer mensaje entrante
	data, err := io.ReadAll(conn)
	if err != nil {
		fmt.Printf("Error leyendo datos: %v\n", err)
		return
	}

	var task models.TaskMessage
	if err := json.Unmarshal(data, &task); err != nil {
		fmt.Printf("Error parseando JSON: %v\n", err)
		return
	}

	// Procesar la tarea
	var resp models.CoordinatorResponse
	switch task.Type {
	case models.RequestRecommendation:
		sims := compute.CosineSimilarityForUser(task.Matrix, task.UserIndex)
		preds := compute.PredictRatings(task.Matrix, sims, task.UserIndex, task.K)
		indexes := compute.SortIndexesByScore(preds)

		resp = models.CoordinatorResponse{
			Result:  preds,
			Indexes: indexes,
		}

	case models.RequestSimilarity:
		simMatrix := compute.CosineSimilarityMatrix(task.Matrix)
		resp = models.CoordinatorResponse{Result: simMatrix}

	default:
		fmt.Printf("Tipo de tarea desconocido: %s\n", task.Type)
		return
	}

	// Enviar respuesta al coordinador
	jsonData, _ := json.Marshal(resp)
	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Printf("Error enviando respuesta: %v\n", err)
		return
	}

	fmt.Println("Tarea completada y enviada al Coordinador")
}

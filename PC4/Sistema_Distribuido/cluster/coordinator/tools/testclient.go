package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sdr/cluster/shared/models"
	"time"
)

func main() {

	msg := models.TaskMessage{
		Type:      models.RequestSimilarity,
		Matrix:    [][]float64{{1, 0.5, 0}, {0.2, 0, 0.8}, {0, 0.9, 0.4}},
		UserIndex: 0,
		K:         2,
	}

	data, _ := json.Marshal(msg)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8081", 5*time.Second)
	if err != nil {
		log.Fatalf("error conectando: %v", err)
	}
	defer conn.Close()

	// ENVÍO DEL MENSAJE
	_, err = conn.Write(data)
	if err != nil {
		log.Fatalf("error escribiendo: %v", err)
	}

	// Cerrar el lado de escritura para que el servidor sepa que ya terminamos
	tcpConn := conn.(*net.TCPConn)
	tcpConn.CloseWrite()

	// LEER RESPUESTA
	resp, err := io.ReadAll(conn)
	if err != nil {
		log.Fatalf("error leyendo respuesta: %v", err)
	}

	fmt.Println("Respuesta cruda:", string(resp))

	var out models.CoordinatorResponse
	if err := json.Unmarshal(resp, &out); err != nil {
		log.Fatalf("error parseando respuesta: %v", err)
	}

	fmt.Printf("CoordinatorResponse: %+v\n", out)
}

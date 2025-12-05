package tcpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sdr/cluster/shared/models"
	"time"
)

// SendTask envía un mensaje JSON al Worker y recibe su respuesta
func SendTask(addr string, task models.TaskMessage) (models.CoordinatorResponse, error) {
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return models.CoordinatorResponse{}, fmt.Errorf("no se pudo conectar al worker %s: %v", addr, err)
	}
	defer conn.Close()

	// Enviar solicitud
	data, _ := json.Marshal(task)
	_, err = conn.Write(data)
	if err != nil {
		return models.CoordinatorResponse{}, fmt.Errorf("error al enviar datos al worker %s: %v", addr, err)
	}

	// Cerrar escritura para indicar fin de mensaje
	conn.(*net.TCPConn).CloseWrite()

	// Leer toda la respuesta completa (evita "unexpected end of JSON input")
	dataResp, err := io.ReadAll(conn)
	if err != nil {
		return models.CoordinatorResponse{}, fmt.Errorf("error al leer respuesta del worker %s: %v", addr, err)
	}

	// Deserializar respuesta JSON
	var resp models.CoordinatorResponse
	if err := json.Unmarshal(dataResp, &resp); err != nil {
		return models.CoordinatorResponse{}, fmt.Errorf("error al parsear respuesta del worker %s: %v", addr, err)
	}

	return resp, nil
}

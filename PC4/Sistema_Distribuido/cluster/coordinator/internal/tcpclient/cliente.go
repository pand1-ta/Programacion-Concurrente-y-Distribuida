package tcpclient

import (
	"encoding/json"
	"fmt"
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

	// cerrar la escritura para que el worker termine de leer
	conn.(*net.TCPConn).CloseWrite()

	// Leer respuesta
	buf := make([]byte, 65536)
	n, err := conn.Read(buf)

	if err != nil {
		return models.CoordinatorResponse{}, fmt.Errorf("error al leer respuesta del worker %s: %v", addr, err)
	}

	var resp models.CoordinatorResponse
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		return models.CoordinatorResponse{}, fmt.Errorf("error al parsear respuesta del worker %s: %v", addr, err)
	}

	return resp, nil
}

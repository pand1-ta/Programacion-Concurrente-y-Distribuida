package tcpserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sdr/cluster/coordinator/internal/dispatcher"
	"sdr/cluster/shared/models"
)

type TCPServer struct {
	Addr string
}

// Run inicia el servidor TCP del coordinador
func (s *TCPServer) Run() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("error escuchando en %s: %w", s.Addr, err)
	}

	log.Printf("Nodo coordinador TCP escuchando en %s", s.Addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("error aceptando conexión: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Println("error leyendo mensaje:", err)
		return
	}

	// Deserializar el TaskMessage
	var msg models.TaskMessage
	err = json.Unmarshal(data, &msg)
	if err != nil {
		log.Println("error parseando JSON:", err)
		return
	}

	log.Printf("El nodo coordinador recibió una solicitud: %s", msg.Type)

	// Llamar al dispatcher para procesar la solicitud
	resp, err := dispatcher.Process(msg)
	if err != nil {
		log.Println("error procesando tarea:", err)
		return
	}

	// Serializar la respuesta
	out, err := json.Marshal(resp)
	if err != nil {
		log.Println("error serializando respuesta:", err)
		return
	}

	// Enviar de vuelta a la API
	// Enviar de vuelta a la API
	_, err = conn.Write(out)
	if err != nil {
		log.Println("error enviando respuesta:", err)
		return
	}

	conn.Close()
	log.Println("Respuesta enviada a la API y conexión cerrada")

}

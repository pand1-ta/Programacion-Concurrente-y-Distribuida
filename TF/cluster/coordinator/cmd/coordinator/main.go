package main

import (
	"fmt"
	"log"
	"os"
	"sdr/cluster/coordinator/internal/tcpserver"
)

func main() {
	port := os.Getenv("COORDINATOR_PORT")
	if port == "" {
		port = "8081"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	srv := &tcpserver.TCPServer{Addr: addr}

	log.Printf("Iniciando coordinador (TCP) en %s", addr)
	if err := srv.Run(); err != nil {
		log.Fatalf("coordinator error: %v", err)
	}
}




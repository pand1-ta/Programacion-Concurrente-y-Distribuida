package coordinator

import (
	"encoding/json"
	"net"
	"time"
)

type CoordinatorRequest struct {
	Type      string      `json:"type"`
	Matrix    [][]float64 `json:"matrix,omitempty"`
	UserIndex int         `json:"userIndex"`
	K         int         `json:"k,omitempty"`
}

type CoordinatorResponse struct {
	Indexes []int `json:"indexes"`
}

type CoordinatorClient struct {
	Addr        string
	DialTimeout time.Duration
}

func NewCoordinatorClient(addr string) *CoordinatorClient {
	return &CoordinatorClient{Addr: addr, DialTimeout: 5 * time.Second}
}

func (c *CoordinatorClient) RequestRecommendations(userIndex int, matrix [][]float64, k int) ([]int, error) {
	conn, err := net.DialTimeout("tcp", c.Addr, c.DialTimeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	req := CoordinatorRequest{Type: "RECOMMENDATION", Matrix: matrix, UserIndex: userIndex, K: k}
	data, _ := json.Marshal(req)
	conn.Write(data)
	conn.(*net.TCPConn).CloseWrite()

	var resp CoordinatorResponse
	dec := json.NewDecoder(conn)
	if err := dec.Decode(&resp); err != nil {
		return nil, err
	}
	return resp.Indexes, nil
}

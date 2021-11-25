package client

import (
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// GetGRPCConn get grpc client conn
func GetGRPCConn(conn string) (*grpc.ClientConn, error) {
	if conn == "" {
		return nil, fmt.Errorf("conn is empty")
	}
	targets := strings.Split(conn, ",")
	for _, target := range targets {
		conn, err := grpc.Dial(target, grpc.WithInsecure())
		if err != nil {
			continue
		}

		connState := conn.GetState()
		if connState != connectivity.Idle && connState != connectivity.Ready {
			continue
		}
		return conn, nil
	}
	return nil, fmt.Errorf("valid conn not found")
}

package client

import (
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	rlk   sync.RWMutex
	conns sync.Map
)

// GetGRPCConn get grpc client conn
func GetGRPCConn(conn string) (*grpc.ClientConn, error) {
	if conn == "" {
		return nil, fmt.Errorf("conn is empty")
	}

	targets := strings.Split(conn, ",")
	for _, target := range targets {
		rlk.RLock()
		_conn, ok := conns.Load(target)
		if ok {
			rlk.RUnlock()
			return _conn.(*grpc.ClientConn), nil
		}
		rlk.RUnlock()

		conn, err := getConn(target)
		if err != nil {
			continue
		}

		conns.Store(target, conn)
		return conn, nil
	}

	return nil, fmt.Errorf("valid conn not found")
}

func getConn(target string) (*grpc.ClientConn, error) {
	rlk.Lock()
	defer rlk.Unlock()

	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	connState := conn.GetState()
	if connState != connectivity.Idle && connState != connectivity.Ready {
		return nil, fmt.Errorf("get conn state not ready: %v", connState)
	}

	return conn, nil
}

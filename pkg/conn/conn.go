package conn

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var target2Conn sync.Map

// GetGRPCConn get grpc client conn
func GetGRPCConn(conn string) (*grpc.ClientConn, error) {
	if conn == "" {
		return nil, fmt.Errorf("conn is empty")
	}
	targets := strings.Split(conn, ",")
	for _, target := range targets {
		v, ok := target2Conn.Load(target)
		if !ok {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(),
				grpc.WithBlock(),
			)
			if err != nil {
				continue
			}
			target2Conn.Store(target, conn)
			return conn, nil
		}

		var conn *grpc.ClientConn
		if _conn, ok := v.(*grpc.ClientConn); ok {
			conn = _conn
		}
		if conn == nil {
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

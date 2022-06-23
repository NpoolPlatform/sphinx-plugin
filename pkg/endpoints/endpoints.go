package endpoints

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

var (
	publicAddrs []string
	localAddrs  []string
)

const (
	AddrSplitter = ","
	AddrMinLen   = 3
)

type Endpoint struct {
	Address string
	IsLocal bool
}

type Manager struct {
	peekRcd []int
}

func NewManager() *Manager {
	if len(localAddrs) > 0 {
		return &Manager{peekRcd: []int{rand.Intn(len(localAddrs))}}
	}
	return &Manager{peekRcd: []int{-rand.Intn(len(publicAddrs)) - 1}}
}

func (endpointmgr *Manager) Peek() (*Endpoint, error) {
	if len(localAddrs) < 1 && len(publicAddrs) < 1 {
		return nil, fmt.Errorf("have no endpoints for plugin")
	}
	endpoint := &Endpoint{}
	currentIdx := endpointmgr.peekRcd[len(endpointmgr.peekRcd)-1]

	if currentIdx >= 0 {
		endpoint.IsLocal = true
		endpoint.Address = localAddrs[currentIdx]
		nextIdx := (currentIdx + 1) % len(localAddrs)
		if len(publicAddrs) < 1 || nextIdx != endpointmgr.peekRcd[0] {
			endpointmgr.peekRcd = append(endpointmgr.peekRcd, nextIdx)
			logger.Sugar().Infof("peek the endpoint: %v", endpoint.Address)
			return endpoint, nil
		}
	} else {
		endpoint.IsLocal = false
		endpoint.Address = publicAddrs[-currentIdx-1]
	}

	nextIdx := rand.Intn(len(publicAddrs))
	endpointmgr.peekRcd = append(endpointmgr.peekRcd, -nextIdx-1)
	logger.Sugar().Infof("peek the endpoint: %v", endpoint.Address)
	return endpoint, nil
}

func init() {
	// read endpoints from env
	_publicAddrs, _ := env.LookupEnv(env.ENVCOINPUBLICAPI)
	if len(_publicAddrs) > AddrMinLen {
		publicAddrs = strings.Split(_publicAddrs, AddrSplitter)
	}

	_localAddrs, _ := env.LookupEnv(env.ENVCOINLOCALAPI)
	if len(_localAddrs) > AddrMinLen {
		localAddrs = strings.Split(_localAddrs, AddrSplitter)
	}

	rand.Seed(time.Now().Unix())
}

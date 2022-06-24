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

type endpoint struct {
	address string
	peeked  bool
}

type Manager struct {
	localEndpoints  []*endpoint
	publicEndpoints []*endpoint
	localCursor     int
	publicCursor    int
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

func NewManager() (*Manager, error) {
	m := Manager{}

	if len(localAddrs) == 0 && len(publicAddrs) == 0 {
		return nil, fmt.Errorf("invalid addresses setting")
	}

	for _, addr := range localAddrs {
		m.localEndpoints = append(m.localEndpoints, &endpoint{
			Address: addr,
			peeked:  false,
		})
	}

	for _, addr := range publicAddrs {
		m.publicEndpoints = append(m.publicEndpoints, &endpoint{
			Address: addr,
			peeked:  false,
		})
	}

	m.localCursor = rand.Intn(len(localAddrs))
	m.publicCursor = rand.Intn(len(publicAddrs))

	return &m
}

func (mgr *Manager) Peek() (string, error) {
	for i := 0; i < len(m.localEndpoints); i++ {
		ep := m.localEndpoints[(i+m.localCursor)%len(m.localEndpoints)]
		if ep.peeked {
			continue
		}
		ep.peeked = true
		m.localCursor = i + 1
		return ep.Address, nil
	}

	for i := 0; i < len(m.publicEnepoints); i++ {
		ep := m.publicEndpoints[(i+m.cursor)%len(m.publicEndpoints)]
		if ep.peeked {
			continue
		}
		ep.peeked = true
		ep.publicCursor = i + 1
		return ep.Address, nil
	}

	return "", fmt.Errorf("invalid endpoint")
}

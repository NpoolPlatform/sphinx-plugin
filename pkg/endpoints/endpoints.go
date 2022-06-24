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
	allAddrs    []string
)

const (
	AddrSplitter = ","
	AddrMinLen   = 3
)

type Manager struct {
	peekOrder     []int
	currentCursor int
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

	allAddrs = append(allAddrs, localAddrs...)
	allAddrs = append(allAddrs, publicAddrs...)

	rand.Seed(time.Now().Unix())
}

func ShuffleOrder(n int) []int {
	if n < 1 {
		return []int{}
	}
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}

	for i := 0; i < n; i++ {
		j := rand.Intn(n)
		order[i], order[j] = order[j], order[i]
	}

	return order
}

func NewManager() (*Manager, error) {
	if len(allAddrs) == 0 {
		return nil, fmt.Errorf("invalid addresses setting")
	}

	localOrder := ShuffleOrder(len(localAddrs))
	publicOrder := ShuffleOrder(len(publicAddrs))
	for i, v := range publicOrder {
		publicOrder[i] = v + len(localAddrs)
	}

	peekOrder := make([]int, 0)
	peekOrder = append(peekOrder, localOrder...)
	peekOrder = append(peekOrder, publicOrder...)

	return &Manager{peekOrder: peekOrder, currentCursor: 0}, nil
}

func (m *Manager) Peek() (addr string, isLocal bool, err error) {
	if m.currentCursor >= len(m.peekOrder) {
		return "", false, fmt.Errorf("fail peek,all endpoints is peeked")
	}

	addr = allAddrs[m.peekOrder[m.currentCursor]]
	m.currentCursor++

	isLocal = false
	if len(localAddrs) != 0 && m.currentCursor < len(localAddrs) {
		isLocal = true
	}

	logger.Sugar().Errorf("peek the endpoint: %v", addr)
	return addr, isLocal, nil
}

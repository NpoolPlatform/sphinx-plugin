package endpoints

import (
	"errors"
	"math/rand"
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
)

var (
	ErrEndpointExhausted = errors.New("all endpoints is peeked")
	ErrEndpointsEmpty    = errors.New("endpoints empty")
)

const (
	AddrSplitter = ","
	AddrMinLen   = 3
)

type Manager struct {
	localAddrs  []string
	publicAddrs []string
}

func NewManager() (*Manager, error) {
	_localAddrs := config.GetENV().LocalWalletAddr
	_publicAddrs := config.GetENV().PublicWalletAddr

	localAddrs := strings.Split(_localAddrs, AddrSplitter)
	publicAddrs := strings.Split(_publicAddrs, AddrSplitter)

	if len(localAddrs) == 0 &&
		len(publicAddrs) == 0 {
		return nil, ErrEndpointsEmpty
	}

	if len(localAddrs) > 1 {
		rand.Shuffle(len(localAddrs), func(i, j int) {
			localAddrs[i], localAddrs[j] = localAddrs[j], localAddrs[i]
		})
	}
	if len(publicAddrs) > 1 {
		rand.Shuffle(len(publicAddrs), func(i, j int) {
			publicAddrs[i], publicAddrs[j] = publicAddrs[j], publicAddrs[i]
		})
	}

	// random start
	return &Manager{
		localAddrs:  localAddrs,
		publicAddrs: publicAddrs,
	}, nil
}

func (m *Manager) Peek() (addr string, err error) {
	ll := len(m.localAddrs)
	pl := len(m.publicAddrs)
	if ll > 0 {
		addr = m.localAddrs[ll-1]
		m.localAddrs = m.localAddrs[0 : ll-1]
		return addr, nil
	}

	if pl > 0 {
		addr = m.publicAddrs[pl-1]
		m.publicAddrs = m.publicAddrs[0 : pl-1]
		return addr, nil
	}

	return "", ErrEndpointExhausted
}

func (m *Manager) Len() int {
	return len(m.localAddrs) + len(m.publicAddrs)
}

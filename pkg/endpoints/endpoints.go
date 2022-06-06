package endpoints

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

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

func Peek(localEndpoint bool) (string, error) {
	var allEndpoints []string
	allEndpoints = append(allEndpoints, localAddrs...)
	if !localEndpoint {
		allEndpoints = append(allEndpoints, publicAddrs...)
	}

	if len(allEndpoints) < 1 {
		return "", fmt.Errorf("have no any endpoints")
	}
	randIndex := rand.Intn(len(allEndpoints))
	return allEndpoints[randIndex], nil
}

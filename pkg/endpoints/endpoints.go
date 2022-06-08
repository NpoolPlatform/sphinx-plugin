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

func Peek(mustLocalEndpoint bool) (endpoint string, isLocal bool, err error) {
	var allEndpoints []string
	var isLocalEndpoint bool
	allEndpoints = append(allEndpoints, localAddrs...)
	if !mustLocalEndpoint {
		allEndpoints = append(allEndpoints, publicAddrs...)
	}

	if len(allEndpoints) < 1 {
		return "", isLocalEndpoint, fmt.Errorf("have no any endpoints")
	}
	randIndex := rand.Intn(len(allEndpoints))
	if randIndex < len(localAddrs) {
		isLocalEndpoint = true
	}

	return allEndpoints[randIndex], isLocalEndpoint, nil
}

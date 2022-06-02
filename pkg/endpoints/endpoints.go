package endpoints

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

var (
<<<<<<< HEAD
	publicAddrs []string
	localAddrs  []string
=======
	publicAddrList []string
	localAddrList  []string
>>>>>>> suport mutple-endpoints for tron
)

const (
	AddrSplitter = ","
	AddrMinLen   = 3
)

func init() {
	// read endpoints from env
	_publicAddrs, _ := env.LookupEnv(env.ENVCOINPUBLICAPI)
	if len(_publicAddrs) > AddrMinLen {
<<<<<<< HEAD
		publicAddrs = strings.Split(_publicAddrs, AddrSplitter)
=======
		publicAddrList = strings.Split(_publicAddrs, AddrDelimiter)
>>>>>>> suport mutple-endpoints for tron
	}

	_localAddrs, _ := env.LookupEnv(env.ENVCOINLOCALAPI)
	if len(_localAddrs) > AddrMinLen {
<<<<<<< HEAD
		localAddrs = strings.Split(_localAddrs, AddrSplitter)
=======
		localAddrList = strings.Split(_localAddrs, AddrDelimiter)
>>>>>>> suport mutple-endpoints for tron
	}

	rand.Seed(time.Now().Unix())
}

<<<<<<< HEAD
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
=======
func Peek(localEndpoint bool) (string, error) {
	var allEndpoints []string
	allEndpoints = append(allEndpoints, localAddrList...)
	if !localEndpoint {
		allEndpoints = append(allEndpoints, publicAddrList...)
	}

	if len(allEndpoints) < 1 {
		return "", fmt.Errorf("have no any endpoints")
	}
	randIndex := rand.Intn(len(allEndpoints))
	return allEndpoints[randIndex], nil
>>>>>>> suport mutple-endpoints for tron
}

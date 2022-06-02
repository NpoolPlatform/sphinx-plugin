package endpoints

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

var (
	publicAddrList []string
	localAddrList  []string
)

const (
	AddrDelimiter = ","
	AddrMinLen    = 2
)

func init() {
	// read endpoints from env
	_publicAddrs, _ := env.LookupEnv(env.ENVCOINPUBLICAPI)
	if len(_publicAddrs) > AddrMinLen {
		publicAddrList = strings.Split(_publicAddrs, AddrDelimiter)
	}

	_localAddrs, _ := env.LookupEnv(env.ENVCOINLOCALAPI)
	if len(_localAddrs) > AddrMinLen {
		localAddrList = strings.Split(_localAddrs, AddrDelimiter)
	}

	rand.Seed(time.Now().Unix())
}

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
}

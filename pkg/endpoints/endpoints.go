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
	thridAddrList []string
	priAddrList   []string
	allAddrList   []string
)

const (
	AddrDelimiter = ","
	AddrMinLen    = 2
)

func Init() {
	// read endpoints from env
	_thridAddrs, _ := env.LookupEnv(env.ENVCOINAPI)
	if len(_thridAddrs) < AddrMinLen {
		thridAddrList = strings.Split(_thridAddrs, AddrDelimiter)
	}

	_priAddrs, _ := env.LookupEnv(env.ENVCOINPRIAPI)
	if len(_priAddrs) < AddrMinLen {
		priAddrList = strings.Split(_priAddrs, AddrDelimiter)
	}

	allAddrList = append(allAddrList, thridAddrList...)
	allAddrList = append(allAddrList, priAddrList...)
	if len(allAddrList) < 1 {
		logger.Sugar().Errorf("fail to read any endpoints from env")
	}

	rand.Seed(time.Now().Unix())
}

func peek(addrs []string) (string, error) {
	if len(addrs) < 1 {
		return "", fmt.Errorf("have no endpoints")
	}
	randIndex := rand.Intn(len(addrs))
	return addrs[randIndex], nil
}

func PeekPri() (string, error) {
	ret, err := peek(priAddrList)
	if err != nil {
		return "", fmt.Errorf("have no private endpoints")
	}
	return ret, nil
}

func PeekThird() (string, error) {
	ret, err := peek(thridAddrList)
	if err != nil {
		return "", fmt.Errorf("have no third endpoints")
	}
	return ret, nil
}

func Peek() (string, error) {
	ret, err := peek(allAddrList)
	if err != nil {
		return "", fmt.Errorf("have no any endpoints")
	}
	return ret, nil
}

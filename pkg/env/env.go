package env

import (
	"errors"
	"os"
)

const (
	// main or test
	ENVCOINNETWORK = "ENV_COIN_NETWORK"

	// FIL BTC ETH
	ENVCOINTYPE = "ENV_COIN_TYPE"

	// fil btc ip:port
	ENVCOINAPI = "ENV_COIN_API"

	// for btc
	ENVCOINUSER = "ENV_COIN_USER"
	ENVCOINPASS = "ENV_COIN_PASS"

	// for fil
	ENVCOINTOKEN = "ENV_COIN_TOKEN"
)

var (
	ErrEVNCoinType    = errors.New("env ENV_COIN_TYPE not found")
	ErrEVNCoinNetwork = errors.New("env ENV_COIN_NETWORK not found")

	ErrENVCoinAPINotFound = errors.New("env ENV_COIN_API not found")

	// btc
	ErrENVCoinUserNotFound = errors.New("env ENV_COIN_USER not found")
	ErrENVCoinPassNotFound = errors.New("env ENV_COIN_PASS not found")

	// fil
	ErrENVCoinTokenNotFound = errors.New("env ENV_COIN_TOKEN not found")
	ErrAddressInvalid       = errors.New("address invalid")
	ErrSignTypeInvalid      = errors.New("sign type invalid")
	ErrFindMsgNotFound      = errors.New("failed to find message")
	ErrCIDInvalid           = errors.New("cid invalid")
)

func LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func CoinInfo() (coinType, networkType string, err error) {
	var ok bool
	coinType, ok = LookupEnv(ENVCOINTYPE)
	if !ok {
		err = ErrEVNCoinType
		return
	}
	networkType, ok = LookupEnv(ENVCOINNETWORK)
	if !ok {
		err = ErrEVNCoinNetwork
		return
	}
	return
}

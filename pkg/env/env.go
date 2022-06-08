package env

import (
	"errors"
	"os"
)

const (
	// main or test
	ENVCOINNET = "ENV_COIN_NET"

	// FIL BTC ETH/USDT SpaceMesh
	ENVCOINTYPE = "ENV_COIN_TYPE"

	// fil btc ip:port
	ENVCOINLOCALAPI  = "ENV_COIN_LOCAL_API"
	ENVCOINPUBLICAPI = "ENV_COIN_PUBLIC_API"

	// for btc
	ENVCOINUSER = "ENV_COIN_USER"
	ENVCOINPASS = "ENV_COIN_PASS"

	// for fil
	ENVCOINTOKEN = "ENV_COIN_TOKEN"

	// for eth/usdt
	ENVCONTRACT = "ENV_CONTRACT"

	// for tron
	ENVCOINJSONRPCLOCALPORT  = "ENV_COIN_JSONRPC_LOCAL_API"
	ENVCOINJSONRPCPUBLICPORT = "ENV_COIN_JSONRPC_PUBLIC_API"
)

var (
	ErrEVNCoinType     = errors.New("env ENV_COIN_TYPE not found")
	ErrEVNCoinNet      = errors.New("env ENV_COIN_NET not found")
	ErrEVNCoinNetValue = errors.New("env ENV_COIN_NET value only support main|test")

	ErrENVCoinLocalAPINotFound  = errors.New("env ENV_COIN_LOCAL_API not found")
	ErrENVCoinPublicAPINotFound = errors.New("env ENV_COIN_PUBLIC_API not found")

	// btc
	ErrENVCoinUserNotFound = errors.New("env ENV_COIN_USER not found")
	ErrENVCoinPassNotFound = errors.New("env ENV_COIN_PASS not found")

	// fil
	ErrENVCoinTokenNotFound = errors.New("env ENV_COIN_TOKEN not found")
	ErrAddressInvalid       = errors.New("address invalid")
	ErrSignTypeInvalid      = errors.New("sign type invalid")
	ErrFindMsgNotFound      = errors.New("failed to find message")
	ErrCIDInvalid           = errors.New("cid invalid")

	// eth/usdt
	ErrENVContractNotFound = errors.New("env ENV_CONTRACT not found")

	// tron
	ErrENVCOINJSONRPCPortNotFound = errors.New("env ENV_COIN_JSONRPC_PORT not found")
	ErrENVCOINGRPCPortNotFound    = errors.New("env ENV_COIN_GRPC_PORT not found")
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
	networkType, ok = LookupEnv(ENVCOINNET)
	if !ok {
		err = ErrEVNCoinNet
		return
	}
	return
}

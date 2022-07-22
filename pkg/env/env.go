package env

import (
	"errors"
	"os"
)

const (
	// main or test
	ENVCOINNET = "ENV_COIN_NET"

	// ENVSYNCINTERVAL sync transaction status on chain interval
	ENVSYNCINTERVAL = "ENV_SYNC_INTERVAL"

	// eg: filecoin/tfilecoin
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
	ENVCOINJSONRPCLOCALAPI  = "ENV_COIN_JSONRPC_LOCAL_API"
	ENVCOINJSONRPCPUBLICAPI = "ENV_COIN_JSONRPC_PUBLIC_API"
)

var (
	// env error----------------------------
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

	// eth/usdt
	ErrENVContractNotFound = errors.New("env ENV_CONTRACT not found")

	// tron
	ErrENVCOINJSONRPCAPINotFound = errors.New("env ENV_COIN_JSONRPC_API not found")
	ErrENVCOINGRPCAPINotFound    = errors.New("env ENV_COIN_GRPC_API not found")

	// not env error----------------------------
	ErrSignTypeInvalid     = errors.New("sign type invalid")
	ErrFindMsgNotFound     = errors.New("failed to find message")
	ErrCIDInvalid          = errors.New("cid invalid")
	ErrAddressInvalid      = errors.New("address invalid")
	ErrAmountInvalid       = errors.New("amount invalid")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrWaitMessageOnChain  = errors.New("wait message on chain")
	ErrContractInvalid     = errors.New("invalid contract address")
	ErrTransactionFail     = errors.New("transaction fail")
)

func LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func CoinInfo() (networkType, coinType string, err error) {
	var ok bool
	networkType, ok = LookupEnv(ENVCOINNET)
	if !ok {
		err = ErrEVNCoinNet
		return
	}

	coinType, ok = LookupEnv(ENVCOINTYPE)
	if !ok {
		err = ErrEVNCoinType
		return
	}
	return
}

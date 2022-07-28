package eth

import (
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

const (
	ETHACCURACY   = 18
	ERC20ACCURACY = 6
	USDCACCURACY  = 6
)

const (
	gasToLow    = `intrinsic gas too low`
	fundsToLow  = `insufficient funds for gas * price + value`
	nonceToLow  = `nonce too low`
	dialTimeout = 3 * time.Second
)

var stopErrMsg = []string{gasToLow, fundsToLow, nonceToLow, env.ErrAddressInvalid.Error()}

// USDTContract ...
var USDTContract = func(chainet int64) string {
	switch chainet {
	case 1:
		return "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	case 1337:
		contract, ok := env.LookupEnv(env.ENVCONTRACT)
		if !ok {
			panic(env.ErrENVContractNotFound)
		}
		return contract
	}
	return ""
}

// USDCContract ...
var USDCContract = func(chainet int64) string {
	switch chainet {
	case 1:
		return "0xa2327a938Febf5FEC13baCFb16Ae10EcBc4cbDCF"
	default:
		contract, ok := env.LookupEnv(env.ENVCONTRACT)
		if !ok {
			panic(env.ErrENVContractNotFound)
		}
		return contract
	}
}

func TxFailErr(err error) bool {
	if err == nil {
		return false
	}

	for _, v := range stopErrMsg {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}

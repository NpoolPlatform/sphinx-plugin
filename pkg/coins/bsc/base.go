package bsc

import (
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

const (
	BNBACCURACY   = 18
	BEP20ACCURACY = 18
)

var (
	gasTooLow   = `intrinsic gas too low`
	fundsTooLow = `insufficient funds for gas * price + value`
	nonceToLow  = `nonce too low`
	stopErrMsg  = []string{gasTooLow, fundsTooLow, nonceToLow}
)

var BUSDContract = func(chainet int64) string {
	switch chainet {
	case 1:
		return "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
	case 1337:
		contract, ok := env.LookupEnv(env.ENVCONTRACT)
		if !ok {
			panic(env.ErrENVContractNotFound)
		}
		return contract
	}
	return ""
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

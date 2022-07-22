package bsc

import (
	"strings"
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

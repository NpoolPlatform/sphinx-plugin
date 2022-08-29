package eth

import (
	"strings"
	"time"
)

const (
	gasToLow    = `intrinsic gas too low`
	fundsToLow  = `insufficient funds for gas * price + value`
	nonceToLow  = `nonce too low`
	dialTimeout = 3 * time.Second
)

var stopErrMsg = []string{gasToLow, fundsToLow, nonceToLow}

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

package fil

import (
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/build"
	"github.com/shopspring/decimal"
)

// FILNetMap fil net map
var FILNetMap = map[string]address.Network{
	coins.CoinNetMain: address.Mainnet,
	coins.CoinNetTest: address.Testnet,
}

var (
	FilTxFaild = `fil tx faild`
	stopErrMsg = []string{FilTxFaild}
)

func SignType(signType string) (crypto.SigType, error) {
	switch signType {
	case "secp256k1":
		return crypto.SigTypeSecp256k1, nil
	case "bls":
		return crypto.SigTypeBLS, nil
	default:
		return crypto.SigTypeUnknown, env.ErrSignTypeInvalid
	}
}

// FIL2AttoFIL not used, at sphinx sign deal
func FIL2AttoFIL(value float64) (float64, bool) {
	return decimal.NewFromFloat(value).
		Mul(decimal.NewFromInt(int64(build.FilecoinPrecision))).
		Float64()
}

func AttoFIL2FIL(value float64) (float64, bool) {
	return decimal.NewFromFloat(value).
		Div(decimal.NewFromInt(int64(build.FilecoinPrecision))).
		Float64()
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

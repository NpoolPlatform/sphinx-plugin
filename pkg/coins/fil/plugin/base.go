package plugin

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/build"
	"github.com/shopspring/decimal"
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

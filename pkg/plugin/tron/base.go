package tron

import (
	"math/big"

	"github.com/shopspring/decimal"
)

const (
	TRC20ACCURACY = 6
	TRXACCURACY   = 6
)

var (
	EmptyTRC20 = big.NewInt(0)
	EmptyTRX   = int64(0)
)

// feeLimit-10^6=1trx
const TRC20FeeLimit int64 = 15000000

func TRC20ToBigInt(value float64) *big.Int {
	return decimal.NewFromFloat(value).Mul(decimal.NewFromBigInt(big.NewInt(1), TRC20ACCURACY)).BigInt()
}

func TRC20ToBigFloat(value *big.Int) *big.Float {
	return decimal.NewFromBigInt(value, 0).Div(decimal.NewFromBigInt(big.NewInt(1), TRC20ACCURACY)).BigFloat()
}

func TRXToInt(value float64) int64 {
	return decimal.NewFromFloat(value).Mul(decimal.NewFromBigInt(big.NewInt(1), TRXACCURACY)).BigInt().Int64()
}

func TRXToBigFloat(value int64) *big.Float {
	return decimal.NewFromInt(value).Div(decimal.NewFromBigInt(big.NewInt(1), TRXACCURACY)).BigFloat()
}

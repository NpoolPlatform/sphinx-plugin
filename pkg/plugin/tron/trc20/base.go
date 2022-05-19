package trc20

import (
	"math/big"

	"github.com/shopspring/decimal"
)

const TRC20ACCURACY = 6

var EmptyInt = big.NewInt(0)

const feeLimit int64 = 5000000

func ToInt(value float64) *big.Int {
	return decimal.NewFromFloat(value).Mul(decimal.NewFromBigInt(big.NewInt(1), TRC20ACCURACY)).BigInt()
}

func ToFloat(value *big.Int) *big.Float {
	return decimal.NewFromBigInt(value, 0).Div(decimal.NewFromBigInt(big.NewInt(1), TRC20ACCURACY)).BigFloat()
}

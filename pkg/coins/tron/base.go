package tron

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/btcsuite/btcutil/base58"
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

var (
	AddressSize            = 42
	AddressPreFixByte byte = 0x41
)

const (
	txExpired        = `Transaction expired`
	fundsToLow       = `balance is not sufficient`
	AddressNotActive = `account not found`
	AddressInvalid   = `address is invalid`
)

var stopErrs = []string{txExpired, fundsToLow, AddressInvalid, AddressNotActive}

var USDTContract = func(chainet string) string {
	switch chainet {
	case coins.CoinNetMain:
		return "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	case coins.CoinNetTest:
		contract, ok := env.LookupEnv(env.ENVCONTRACT)
		if !ok {
			panic(env.ErrENVContractNotFound)
		}
		return contract
	}
	return ""
}

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

func ValidAddress(input string) error {
	var address []byte
	var err error
	if len(input) == AddressSize {
		address, err = fromHexString(input)
	} else if len(input) == 34 {
		address, err = decodeFromBase58Check(input)
	} else if len(input) == 28 {
		address, err = base64.StdEncoding.DecodeString(input)
	} else {
		return env.ErrAddressInvalid
	}

	if err == nil {
		err = validFormat(address)
	}

	return err
}

func validFormat(address []byte) error {
	if len(address) == 0 {
		return env.ErrAddressInvalid
	}
	if len(address) != AddressSize/2 {
		return fmt.Errorf("address length need %v but %v", AddressSize, len(address))
	}
	if address[0] != AddressPreFixByte {
		return fmt.Errorf("address need prefix with %v but %v", AddressPreFixByte, address[0])
	}
	return nil
}

func fromHexString(input string) ([]byte, error) {
	if input == "" {
		return nil, env.ErrAddressInvalid
	}
	input = strings.TrimPrefix(input, "0x")
	if len(input)%2 != 0 {
		input = "0" + input
	}
	return hex.DecodeString(input)
}

func decode58Check(input string) []byte {
	decodeCheck := base58.Decode(input)
	if len(decodeCheck) <= 4 {
		return nil
	}
	decodeData := make([]byte, len(decodeCheck)-4)
	copy(decodeData, decodeCheck)
	hash0 := sha256.Sum256(decodeData)
	hash1 := sha256.Sum256(hash0[:])

	if bytes.Equal(hash1[:4], decodeCheck[len(decodeCheck)-4:]) {
		return decodeData
	}
	return nil
}

func decodeFromBase58Check(input string) ([]byte, error) {
	if input == "" {
		return nil, env.ErrAddressInvalid
	}
	address := decode58Check(input)
	if address == nil {
		return nil, env.ErrAddressInvalid
	}
	if err := validFormat(address); err != nil {
		return nil, env.ErrAddressInvalid
	}
	return address, nil
}

func TxFailErr(err error) bool {
	if err == nil {
		return false
	}

	for _, v := range stopErrs {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}

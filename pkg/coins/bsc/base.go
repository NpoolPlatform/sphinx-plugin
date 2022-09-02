package bsc

import (
	"strings"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
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

	BUSDContract = func(chainet int64) string {
		switch chainet {
		case 56:
			return "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
		case 97:
			contract, ok := env.LookupEnv(env.ENVCONTRACT)
			if !ok {
				panic(env.ErrENVContractNotFound)
			}
			return contract
		}
		return ""
	}

	bscTokenList = []*coins.TokenInfo{
		{OfficialName: "BSC", Decimal: 18, Unit: "BNB", Name: "binancecoin", OfficialContract: "bsc", TokenType: coins.Binancecoin, CoinType: sphinxplugin.CoinType_CoinTypebinancecoin},
		{OfficialName: "BUSD Token", Decimal: 18, Unit: "BUSD", Name: "binanceusd", OfficialContract: "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56", TokenType: coins.Bep20, CoinType: sphinxplugin.CoinType_CoinTypebinanceusd},
	}
)

func init() {
	for _, token := range bscTokenList {
		token.Waight = 100
		token.Net = coins.CoinNetMain
		token.Contract = token.OfficialContract
		register.RegisteTokenInfo(token)
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

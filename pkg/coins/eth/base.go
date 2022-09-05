package eth

import (
	"strings"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
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

func init() {
	for i := range ethTokens {
		ethTokens[i].Net = coins.CoinNetMain
		ethTokens[i].Contract = ethTokens[i].OfficialContract
		ethTokens[i].CoinType = sphinxplugin.CoinType_CoinTypeethereum
		register.RegisteTokenInfo(&ethTokens[i])
	}
}

var ethTokens = []coins.TokenInfo{
	{Waight: 100, OfficialName: "Ethereum", Decimal: 18, Unit: "ETH", Name: "ethereum", TokenType: coins.Ethereum, OfficialContract: "ethereum"},
	{Waight: 100, OfficialName: "Tether USD", Decimal: 6, Unit: "USDT", Name: "usdterc20", TokenType: coins.Erc20, OfficialContract: "0xdAC17F958D2ee523a2206206994597C13D831ec7"},
	{Waight: 100, OfficialName: "Coins USD", Decimal: 6, Unit: "USDC", Name: "usdcerc20", TokenType: coins.Erc20, OfficialContract: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
}

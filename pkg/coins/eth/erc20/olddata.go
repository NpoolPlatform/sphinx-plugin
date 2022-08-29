package erc20

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
)

func init() {
	for i := range olderc20s {
		olderc20s[i].Net = coins.CoinNetMain
		olderc20s[i].Contract = olderc20s[i].OfficialContract
		olderc20s[i].CoinType = sphinxplugin.CoinType_CoinTypeethereum
		register.RegisteTokenInfo(&olderc20s[i])
	}
}

var olderc20s = []coins.TokenInfo{
	{Waight: 100, OfficialName: "Ethereum", Decimal: 6, Unit: "ETH", Name: "ethereum", OfficialContract: "555", TokenType: coins.Ethereum},
	{Waight: 100, OfficialName: "Coins USD", Decimal: 6, OfficialContract: "123", TokenType: coins.Erc20, Unit: "USD", Name: "usdterc20", CoinType: sphinxplugin.CoinType_CoinTypeethereum},
}

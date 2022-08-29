package erc20

import (
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/utils"
)

func init() {
	for i := range erc20s {
		erc20s[i].TokenType = coins.Erc20
		erc20s[i].Net = coins.CoinNetMain
		erc20s[i].Waight = 1
		erc20s[i].Contract = erc20s[i].OfficialContract
		erc20s[i].CoinType = sphinxplugin.CoinType_CoinTypeethereum
		erc20s[i].Name = coins.GenerateName(utils.ToCoinName(erc20s[i].CoinType), erc20s[i].TokenType, erc20s[i].OfficialName)
		register.RegisteTokenInfo(&erc20s[i])
	}
}

var erc20s = []coins.TokenInfo{
	{OfficialName: "Tether USD", Decimal: 6, Unit: "USD", Name: "eth-tether-usd", OfficialContract: "123"},
	{OfficialName: "Coins USD", Decimal: 6, Unit: "USD", Name: "eth-coins-usdc", OfficialContract: "456"},
	{OfficialName: "Ethereum", Decimal: 6, Unit: "ETHh", Name: "ethereum", OfficialContract: "555"},
}

package erc20

import (
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
)

func init() {
	for i := range erc20List {
		erc20List[i].TokenType = "erc20"
		erc20List[i].Net = "main"
		erc20List[i].Waight = 1
		erc20List[i].Contract = erc20List[i].OfficialContract
		erc20List[i].CoinType = sphinxplugin.CoinType_CoinTypeethereum
		erc20List[i].Name = coins.GenerateName(&erc20List[i])
		register.RegisteTokenInfo(&erc20List[i])
	}
}

var erc20List = []coins.TokenInfo{
	{OfficialName: "Dogelon", Decimal: 18, Unit: "ELON", OfficialContract: "0x761d38e5ddf6ccf6cf7c55759d5210750b5d60f3"},
	{OfficialName: "Wootrade Network", Decimal: 18, Unit: "WOO", OfficialContract: "0x4691937a7508860f876c9c0a2a617e7d9e945d4b"},
	{OfficialName: "MXCToken", Decimal: 18, Unit: "MXC", OfficialContract: "0x5ca381bbfb58f0092df149bd3d243b08b9a8386e"},
	{OfficialName: "Swipe", Decimal: 18, Unit: "SXP", OfficialContract: "0x8ce9137d39326ad0cd6491fb5cc0cba0e089b6a9"},
	{OfficialName: "WAX Token", Decimal: 8, Unit: "WAX", OfficialContract: "0x39bb259f66e1c59d5abef88375979b4d20d98022"},
	{OfficialName: "IOSToken", Decimal: 18, Unit: "IOST", OfficialContract: "0xfa1a856cfa3409cfa145fa4e20eb270df3eb21ab"},
	{OfficialName: "Ethereum Name Service", Decimal: 18, Unit: "ENS", OfficialContract: "0xc18360217d8f7ab5e7c516566761ea12ce7f9d72"},
	{OfficialName: "ZRX", Decimal: 18, Unit: "ZRX", OfficialContract: "0xe41d2489571d322189246dafa5ebde1f4699f498"},
	{OfficialName: "ZEON", Decimal: 18, Unit: "ZEON", OfficialContract: "0xe5b826ca2ca02f09c1725e9bd98d9a8874c30532"},
	{OfficialName: "IoTeX Network", Decimal: 18, Unit: "IOTX", OfficialContract: "0x6fb3e0a217407efff7ca062d46c26e5d60a14d69"},
	{OfficialName: "WQtum", Decimal: 18, Unit: "WQTUM", OfficialContract: "0x3103df8f05c4d8af16fd22ae63e406b97fec6938"},
	{OfficialName: "LoopringCoin V2", Decimal: 18, Unit: "LRC", OfficialContract: "0xbbbbca6a901c926f240b89eacb641d8aec7aeafd"},
	{OfficialName: "Celsius", Decimal: 4, Unit: "CEL", OfficialContract: "0xaaaebe6fe48e54f431b0c390cfaf0b017d09d42d"},
	{OfficialName: "Zilliqa", Decimal: 12, Unit: "ZIL", OfficialContract: "0x05f4a42e251f2d52b8ed15e9fedaacfcef1fad27"},
	{OfficialName: "HuobiToken", Decimal: 18, Unit: "HT", OfficialContract: "0x6f259637dcd74c767781e37bc6133cd6a68aa161"},
	{OfficialName: "Graph Token", Decimal: 18, Unit: "GRT", OfficialContract: "0xc944e90c64b2c07662a292be6244bdf05cda44a7"},
	{OfficialName: "KuCoin Token", Decimal: 6, Unit: "KCS", OfficialContract: "0xf34960d9d60be18cc1d5afc1a6f012a723a28811"},
	{OfficialName: "chiliZ", Decimal: 18, Unit: "CHZ", OfficialContract: "0x3506424f91fd33084466f402d5d97f05f8e3b4af"},
	{OfficialName: "ApeCoin", Decimal: 18, Unit: "APE", OfficialContract: "0x4d224452801aced8b2f0aebe155379bb5d594381"},
	{OfficialName: "Chain", Decimal: 18, Unit: "XCN", OfficialContract: "0xa2cd3d43c775978a96bdbf12d733d5a1ed94fb18"},
	{OfficialName: "ChainLink Token", Decimal: 18, Unit: "LINK", OfficialContract: "0x514910771af9ca656af840dff83e8264ecf986ca"},
	{OfficialName: "Matic Token", Decimal: 18, Unit: "MATIC", OfficialContract: "0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0"},
	{OfficialName: "BNB", Decimal: 18, Unit: "BNB", OfficialContract: "0xB8c77482e45F1F44dE1745F52C74426C631bDD52"},
	{OfficialName: "Tether USD", Decimal: 6, Unit: "USDT", OfficialContract: "0xdac17f958d2ee523a2206206994597c13d831ec7"},
}

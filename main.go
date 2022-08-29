package main

import (
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/getter"
)

func main() {
	fmt.Println(getter.GetTokenInfos(sphinxplugin.CoinType_CoinTypetethereum))
}

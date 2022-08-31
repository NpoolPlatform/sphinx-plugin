package erc20

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/eth"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
)

func init() {
	register.RegisteTokenHandler(
		coins.Erc20,
		register.OpWalletNew,
		eth.CreateEthAccount,
	)
	register.RegisteTokenHandler(
		coins.Erc20,
		register.OpSign,
		eth.EthMsg,
	)
}

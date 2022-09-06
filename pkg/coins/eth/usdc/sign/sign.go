package sign

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	eth_sign "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/eth"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
)

func init() {
	// main
	register.RegisteTokenHandler(
		coins.USDC,
		register.OpWalletNew,
		eth_sign.CreateEthAccount,
	)
	register.RegisteTokenHandler(
		coins.USDC,
		register.OpSign,
		eth_sign.Msg,
	)
}

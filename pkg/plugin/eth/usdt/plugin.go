package usdt

import (
	"context"
	"math/big"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/eth"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// WalletBalance ..
func WalletBalance(ctx context.Context, addr string) (*big.Int, error) {
	client, err := eth.Client()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	tetherERC20Token, err := NewTetherToken(common.HexToAddress(plugin.USDTContractID(chainID.Int64())), client)
	if err != nil {
		return nil, err
	}

	return tetherERC20Token.BalanceOf(
		&bind.CallOpts{
			Pending: true,
			Context: ctx,
		},
		common.HexToAddress(addr),
	)
}

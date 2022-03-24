package usdt

import (
	"context"
	"math/big"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/eth"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type BigUSDT struct {
	Decimal *big.Int
	Balance *big.Int
}

// WalletBalance ..
func WalletBalance(ctx context.Context, addr string) (*BigUSDT, error) {
	client, err := eth.Client()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(addr) {
		return nil, eth.ErrAddrNotValid
	}

	tetherERC20Token, err := NewTetherToken(common.HexToAddress(plugin.USDTContractID(chainID.Int64())), client)
	if err != nil {
		return nil, err
	}

	decimal, err := tetherERC20Token.Decimals(&bind.CallOpts{
		Pending: true,
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	balance, err := tetherERC20Token.BalanceOf(
		&bind.CallOpts{
			Pending: true,
			Context: ctx,
		},
		common.HexToAddress(addr),
	)
	if err != nil {
		return nil, err
	}

	return &BigUSDT{
		Decimal: decimal,
		Balance: balance,
	}, nil
}

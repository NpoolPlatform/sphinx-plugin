package usdt

import (
	"context"
	"math/big"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/eth"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BigUSDT struct {
	Decimal *big.Int
	Balance *big.Int
}

func ERC20Balance(ctx context.Context, addr string, client *ethclient.Client) (*BigUSDT, error) {
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(addr) {
		return nil, eth.ErrAddrNotValid
	}

	tetherERC20Token, err := NewTetherToken(common.HexToAddress(plugin.USDTContract(chainID.Int64())), client)
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

// WalletBalance ..
func WalletBalance(ctx context.Context, addr string) (*BigUSDT, error) {
	eClient := eth.Client()
	var client *ethclient.Client
	var err error
	var ret *BigUSDT
	localEndpoint := true
	for i := 0; i < eth.MaxRetries; i++ {
		client, err = eClient.GetNode(localEndpoint)
		localEndpoint = false
		if err != nil {
			continue
		}

		ret, err = ERC20Balance(ctx, addr, client)
		if err == nil {
			break
		}
	}
	return ret, err
}

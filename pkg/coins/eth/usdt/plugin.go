package usdt

import (
	"context"
	"fmt"
	"math/big"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
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

	tetherERC20Token, err := NewTetherToken(common.HexToAddress(coins.USDTContract(chainID.Int64())), client)
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

	var err error
	var ret *BigUSDT
	err = eClient.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		syncRet, _err := c.SyncProgress(ctx)
		if _err != nil {
			return true, _err
		}
		if syncRet != nil {
			return true, fmt.Errorf(
				"node is syncing ,current block %v ,highest block %v ",
				syncRet.CurrentBlock, syncRet.HighestBlock,
			)
		}

		ret, err = ERC20Balance(ctx, addr, c)
		if err == nil && ret != nil {
			return false, nil
		}
		return true, err
	})
	if ret == nil {
		return nil, fmt.Errorf("get erc20balance faild,%v", err)
	}
	return ret, err
}

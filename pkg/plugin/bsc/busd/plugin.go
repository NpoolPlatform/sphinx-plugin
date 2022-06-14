package busd

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/bsc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrContractAddrInvalid = errors.New("contract address is invalid")
	ErrAccountAddrInvalid  = errors.New("account address is invalid")
)

func Bep20Balance(ctx context.Context, addr string, client bind.ContractBackend) (*big.Int, error) {
	contract := config.GetENV().Contract
	if !common.IsHexAddress(contract) {
		return nil, ErrContractAddrInvalid
	}

	if !common.IsHexAddress(addr) {
		return nil, ErrAccountAddrInvalid
	}

	usdt, err := NewBusd(common.HexToAddress(contract), client)
	if err != nil {
		return nil, err
	}

	return usdt.BalanceOf(&bind.CallOpts{
		Pending: true,
		Context: ctx,
	}, common.HexToAddress(addr))
}

func WalletBalance(ctx context.Context, addr string) (*big.Int, error) {
	var ret *big.Int
	var err error
	var client *bsc.BClients
	client, err = bsc.Client()
	if err != nil {
		return nil, err
	}
	for i := 0; i < bsc.MaxRetryNum; i++ {
		err = client.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) error {
			syncRet, err := c.SyncProgress(ctx)
			if err != nil {
				return err
			}
			if syncRet != nil && syncRet.CurrentBlock < syncRet.HighestBlock {
				return fmt.Errorf(
					"node is syncing ,current block %v ,highest block %v ",
					syncRet.CurrentBlock, syncRet.HighestBlock,
				)
			}
			ret, err = Bep20Balance(ctx, addr, c)
			return err
		})
		if err == nil {
			return ret, nil
		}
	}
	return ret, err
}

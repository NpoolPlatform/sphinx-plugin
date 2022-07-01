package plugin

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
	bsc_plugin "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// here register plugin func
func init() {
	// main
	// coins.RegisterBalance(
	// 	sphinxplugin.CoinType_CoinTypebinancecoin,
	// 	sphinxproxy.TransactionType_Balance,
	// 	WalletBalance,
	// )
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionState_TransactionStateWait,
		bsc_plugin.PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		bsc_plugin.SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionState_TransactionStateSync,
		bsc_plugin.SyncTxState,
	)

	// test
	// coins.RegisterBalance(
	// 	sphinxplugin.CoinType_CoinTypetbinancecoin,
	// 	sphinxproxy.TransactionType_Balance,
	// 	WalletBalance,
	// )
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinancecoin,
		sphinxproxy.TransactionState_TransactionStateWait,
		bsc_plugin.PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinancecoin,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		bsc_plugin.SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinancecoin,
		sphinxproxy.TransactionState_TransactionStateSync,
		bsc_plugin.SyncTxState,
	)
}

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

	usdt, err := NewBEP20Token(common.HexToAddress(contract), client)
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
	client := bsc.Client()
	err = client.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		syncRet, err := c.SyncProgress(ctx)
		if err != nil {
			return true, err
		}
		if syncRet != nil && syncRet.CurrentBlock < syncRet.HighestBlock {
			return true, fmt.Errorf(
				"node is syncing ,current block %v ,highest block %v ",
				syncRet.CurrentBlock, syncRet.HighestBlock,
			)
		}
		ret, err = Bep20Balance(ctx, addr, c)

		return true, err
	})
	return ret, err
}

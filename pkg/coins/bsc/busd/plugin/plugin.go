package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
	bsc_plugin "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	plugin_types "github.com/NpoolPlatform/sphinx-plugin/pkg/types"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// here register plugin func
func init() {
	// main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypebinanceusd,
		sphinxproxy.TransactionType_Balance,
		WalletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinanceusd,
		sphinxproxy.TransactionState_TransactionStateWait,
		bsc_plugin.PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinanceusd,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		bsc_plugin.SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinanceusd,
		sphinxproxy.TransactionState_TransactionStateSync,
		bsc_plugin.SyncTxState,
	)

	// testTransactionStateWait
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetbinanceusd,
		sphinxproxy.TransactionType_Balance,
		WalletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinanceusd,
		sphinxproxy.TransactionState_TransactionStateWait,
		bsc_plugin.PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinanceusd,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		bsc_plugin.SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinanceusd,
		sphinxproxy.TransactionState_TransactionStateSync,
		bsc_plugin.SyncTxState,
	)
	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypebinanceusd, bsc.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetbinanceusd, bsc.TxFailErr)
	if err != nil {
		panic(err)
	}
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

func walletBalance(ctx context.Context, addr string) (*big.Int, error) {
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

func WalletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	wbReq := &plugin_types.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return nil, err
	}

	bl, err := walletBalance(ctx, wbReq.Address)
	if err != nil {
		return nil, err
	}

	balance, ok := big.NewFloat(0).SetString(bl.String())
	if !ok {
		return nil, errors.New("convert balance string to float64 error")
	}

	balance.Quo(balance, big.NewFloat(math.Pow10(bsc.BEP20ACCURACY)))
	f, exact := balance.Float64()
	if exact != big.Exact {
		log.Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
	}

	wbResp := &plugin_types.WalletBalanceResponse{
		Balance:    f,
		BalanceStr: balance.String(),
	}
	out, err = json.Marshal(wbResp)

	return out, err
}

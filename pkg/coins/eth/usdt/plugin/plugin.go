package usdt

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
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	eth_plugin "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/usdt"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	plugin_types "github.com/NpoolPlatform/sphinx-plugin/pkg/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// here register plugin func
func init() {
	// main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypeusdterc20,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdterc20,
		sphinxproxy.TransactionState_TransactionStateWait,
		eth_plugin.PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdterc20,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		eth_plugin.SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdterc20,
		sphinxproxy.TransactionState_TransactionStateSync,
		eth_plugin.SyncTxState,
	)

	// test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetusdterc20,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdterc20,
		sphinxproxy.TransactionState_TransactionStateWait,
		eth_plugin.PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdterc20,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		eth_plugin.SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdterc20,
		sphinxproxy.TransactionState_TransactionStateSync,
		eth_plugin.SyncTxState,
	)

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypeusdterc20, eth.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetusdterc20, eth.TxFailErr)
	if err != nil {
		panic(err)
	}
}

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
		return nil, env.ErrAddressInvalid
	}

	tetherERC20Token, err := usdt.NewTetherToken(common.HexToAddress(eth.USDTContract(chainID.Int64())), client)
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

// walletBalance ..
func _walletBalance(ctx context.Context, addr string) (*BigUSDT, error) {
	eClient := eth.Client()

	var err error
	var ret *BigUSDT
	err = eClient.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
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

func walletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	wbReq := &plugin_types.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return nil, err
	}

	bl, err := _walletBalance(ctx, wbReq.Address)
	if err != nil {
		return nil, err
	}

	balance, ok := big.NewFloat(0).SetString(bl.Balance.String())
	if !ok {
		return nil, errors.New("convert balance string to float64 error")
	}

	balance.Quo(balance, big.NewFloat(math.Pow10(eth.ERC20ACCURACY)))
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

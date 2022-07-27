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
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	eth_plugin "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/plugin"
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
		sphinxplugin.CoinType_CoinTypeusdcerc20,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdcerc20,
		sphinxproxy.TransactionState_TransactionStateWait,
		eth_plugin.PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdcerc20,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		eth_plugin.SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdcerc20,
		sphinxproxy.TransactionState_TransactionStateSync,
		eth_plugin.SyncTxState,
	)

	// test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetusdcerc20,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdcerc20,
		sphinxproxy.TransactionState_TransactionStateWait,
		eth_plugin.PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdcerc20,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		eth_plugin.SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdcerc20,
		sphinxproxy.TransactionState_TransactionStateSync,
		eth_plugin.SyncTxState,
	)

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypeusdcerc20, eth.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetusdcerc20, eth.TxFailErr)
	if err != nil {
		panic(err)
	}
}

type BigUSDC struct {
	Decimal uint8
	Balance *big.Int
}

func USDCBalance(ctx context.Context, addr string, client *ethclient.Client) (*BigUSDC, error) {
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{
		Pending: true,
		Context: ctx,
	}

	usdcContract := eth.USDCContract(chainID.Int64())
	usdcProxyAddr := common.HexToAddress(usdcContract)
	usdcProxy, err := NewUsdcProxy(usdcProxyAddr, client)
	if err != nil {
		return nil, err
	}

	usdcAddr, err := usdcProxy.Implementation(callOpts)
	if err != nil {
		return nil, err
	}

	usdcImpl, err := NewUsdc(usdcAddr, client)
	if err != nil {
		return nil, err
	}

	decimal, err := usdcImpl.Decimals(callOpts)
	if err != nil {
		return nil, err
	}

	balance, err := usdcImpl.BalanceOf(
		callOpts,
		common.HexToAddress(addr),
	)
	if err != nil {
		return nil, err
	}

	return &BigUSDC{
		Decimal: decimal,
		Balance: balance,
	}, nil
}

func walletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	wbReq := &plugin_types.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)

	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(wbReq.Address) {
		return nil, env.ErrAddressInvalid
	}

	eClient := eth.Client()
	var bl *BigUSDC

	err = eClient.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		bl, err = USDCBalance(ctx, wbReq.Address, c)
		if err != nil || bl == nil {
			return true, err
		}
		return false, err
	})

	if bl == nil || err != nil {
		return nil, fmt.Errorf("get erc20balance faild,%v", err)
	}

	balance, ok := big.NewFloat(0).SetString(bl.Balance.String())
	if !ok {
		return nil, errors.New("convert balance string to float64 error")
	}

	balance.Quo(balance, big.NewFloat(math.Pow10(eth.USDCACCURACY)))
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

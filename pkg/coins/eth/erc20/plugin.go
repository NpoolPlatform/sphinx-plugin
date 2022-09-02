package erc20

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	eth_plugin "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/eth"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	plugin_types "github.com/NpoolPlatform/sphinx-plugin/pkg/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// here register plugin func
func init() {
	register.RegisteTokenHandler(
		coins.Erc20,
		register.OpGetBalance,
		walletBalance,
	)
	register.RegisteTokenHandler(
		coins.Erc20,
		register.OpPreSign,
		PreSign,
	)
	register.RegisteTokenHandler(
		coins.Erc20,
		register.OpBroadcast,
		eth_plugin.SendRawTransaction,
	)
	register.RegisteTokenHandler(
		coins.Erc20,
		register.OpSyncTx,
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

func walletBalance(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	wbReq := &plugin_types.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)

	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(wbReq.Address) {
		return nil, env.ErrAddressInvalid
	}

	eClient := eth.Client()
	var bl *big.Int

	err = eClient.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		callOpts := &bind.CallOpts{
			Pending: true,
			Context: ctx,
		}

		usdcImpl, err := NewErc20token(common.HexToAddress(tokenInfo.Contract), c)
		if err != nil {
			return true, err
		}

		bl, err = usdcImpl.BalanceOf(
			callOpts,
			common.HexToAddress(wbReq.Address),
		)
		if err != nil {
			return true, err
		}
		return false, err
	})

	if bl == nil || err != nil {
		return nil, fmt.Errorf("get erc20balance faild,%v", err)
	}

	balance, ok := big.NewFloat(0).SetString(bl.String())
	if !ok {
		return nil, errors.New("convert balance string to float64 error")
	}

	balance.Quo(balance, big.NewFloat(math.Pow10(tokenInfo.Decimal)))
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

func PreSign(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	baseInfo := &plugin_types.BaseInfo{}
	err = json.Unmarshal(in, baseInfo)
	if err != nil {
		return nil, err
	}

	if !coins.CheckSupportNet(baseInfo.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	if !common.IsHexAddress(baseInfo.From) {
		return nil, env.ErrAddressInvalid
	}

	client := eth.Client()

	var (
		chainID  *big.Int
		nonce    uint64
		gasPrice *big.Int
		gasLimit = uint64(300_000)
	)

	err = client.WithClient(ctx, func(ctx context.Context, cli *ethclient.Client) (bool, error) {
		chainID, err = cli.NetworkID(ctx)
		if err != nil || chainID == nil {
			return true, err
		}

		nonce, err = cli.PendingNonceAt(ctx, common.HexToAddress(baseInfo.From))
		if err != nil {
			return true, err
		}

		gasPrice, err = cli.SuggestGasPrice(ctx)
		if err != nil || gasPrice == nil {
			return true, err
		}

		return false, err
	})
	if err != nil {
		return nil, err
	}

	amount := big.NewFloat(baseInfo.Value)
	amount.Mul(amount, big.NewFloat(math.Pow10(tokenInfo.Decimal)))

	amountBig, ok := big.NewInt(0).SetString(amount.Text('f', 0), 10)
	if !ok {
		return in, errors.New("invalid usd amount")
	}

	_abi, err := Erc20tokenMetaData.GetAbi()
	input, err := _abi.Pack(
		"transfer",
		common.HexToAddress(baseInfo.To),
		amountBig,
	)
	if err != nil {
		return in, err
	}

	if amountBig.Cmp(common.Big0) <= 0 {
		return nil, errors.New("invalid eth amount")
	}

	// build tx
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(tokenInfo.OfficialContract),
		big.NewInt(0),
		gasLimit,
		big.NewInt(gasPrice.Int64()),
		input,
	)

	fmt.Println(nonce,
		common.HexToAddress(tokenInfo.OfficialContract),
		big.NewInt(0),
		gasLimit,
		big.NewInt(gasPrice.Int64()),
		input)

	info := &eth.PreSignData{
		ChainID: chainID,
		From:    baseInfo.From,
		Tx:      tx,
	}

	out, err = json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return out, err
}
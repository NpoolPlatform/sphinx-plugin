package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
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

const USDCACCURACY = 6

// here register plugin func
func init() {
	register.RegisteTokenHandler(
		coins.USDC,
		register.OpGetBalance,
		walletBalance,
	)
	register.RegisteTokenHandler(
		coins.USDC,
		register.OpPreSign,
		preSign,
	)
	register.RegisteTokenHandler(
		coins.USDC,
		register.OpBroadcast,
		eth_plugin.SendRawTransaction,
	)
	register.RegisteTokenHandler(
		coins.USDC,
		register.OpSyncTx,
		eth_plugin.SyncTxState,
	)

	err := register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypeusdcerc20, bsc.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypetusdcerc20, bsc.TxFailErr)
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

	usdcContract := USDCContract(chainID.Int64())
	callOpts := &bind.CallOpts{
		Pending: true,
		Context: ctx,
	}

	usdcAddr := common.HexToAddress(usdcContract)
	usdcImpl, err := NewUsdcv21(usdcAddr, client)
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

func walletBalance(ctx context.Context, in []byte, token *coins.TokenInfo) (out []byte, err error) {
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
		return nil, fmt.Errorf("get erc20balance failed,%v", err)
	}

	balance, ok := big.NewFloat(0).SetString(bl.Balance.String())
	if !ok {
		return nil, errors.New("convert balance string to float64 error")
	}

	balance.Quo(balance, big.NewFloat(math.Pow10(USDCACCURACY)))
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

// USDCContract ...
var USDCContract = func(chainet int64) string {
	switch chainet {
	case 1:
		return "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	default:
		contract, ok := env.LookupEnv(env.ENVCONTRACT)
		if !ok {
			panic(env.ErrENVContractNotFound)
		}
		return contract
	}
}

func preSign(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
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

	tokenInfo.Contract = USDCContract(chainID.Int64())
	if !common.IsHexAddress(tokenInfo.Contract) && common.HexToAddress("") != common.HexToAddress(tokenInfo.Contract) {
		// TODO:is not env error,it will be replaced
		return nil, env.ErrContractInvalid
	}

	amount := big.NewFloat(baseInfo.Value)
	amount.Mul(amount, big.NewFloat(math.Pow10(tokenInfo.Decimal)))

	amountBig, ok := big.NewInt(0).SetString(amount.Text('f', 0), 10)
	if !ok {
		return in, errors.New("invalid usd amount")
	}

	_abi, err := Usdcv21MetaData.GetAbi()
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
		common.HexToAddress(tokenInfo.Contract),
		big.NewInt(0),
		gasLimit,
		big.NewInt(gasPrice.Int64()),
		input,
	)

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

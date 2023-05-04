package eth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/rlp"
)

// here register plugin func
func init() {
	register.RegisteTokenHandler(
		coins.Ethereum,
		register.OpGetBalance,
		walletBalance,
	)
	register.RegisteTokenHandler(
		coins.Ethereum,
		register.OpPreSign,
		PreSign,
	)
	register.RegisteTokenHandler(
		coins.Ethereum,
		register.OpBroadcast,
		SendRawTransaction,
	)
	register.RegisteTokenHandler(
		coins.Ethereum,
		register.OpSyncTx,
		SyncTxState,
	)

	err := register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypeethereum, eth.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypetethereum, eth.TxFailErr)
	if err != nil {
		panic(err)
	}
}

func walletBalance(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	wbReq := &ct.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(wbReq.Address) {
		return nil, env.ErrAddressInvalid
	}

	client := eth.Client()

	var bl *big.Int
	err = client.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		bl, err = c.BalanceAt(ctx, common.HexToAddress(wbReq.Address), nil)
		if err == nil && bl != nil {
			return false, err
		}
		if err != nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	balance := eth.ToEth(bl)

	wbResp := &ct.WalletBalanceResponse{
		Balance:    balance.InexactFloat64(),
		BalanceStr: balance.String(),
	}
	out, err = json.Marshal(wbResp)

	return out, err
}

func PreSign(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	baseInfo := &ct.BaseInfo{}
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
		bl       *big.Int
		gasPrice *big.Int
		gasLimit = uint64(21_000)
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

		bl, err = cli.BalanceAt(ctx, common.HexToAddress(baseInfo.From), nil)
		if err == nil && bl != nil {
			return false, err
		}
		if err != nil {
			return true, err
		}

		return false, err
	})

	if err != nil {
		return nil, err
	}

	amountBig, ok := eth.ToWei(baseInfo.Value)
	if !ok {
		return nil, errors.New(eth.AmountInvalid)
	}

	if amountBig.Cmp(common.Big0) <= 0 {
		return nil, errors.New(eth.AmountInvalid)
	}

	gasLimitBig := big.NewInt(int64(gasLimit))
	totalGas := big.NewInt(0).Mul(gasPrice, gasLimitBig)

	totalFunds := big.NewInt(0).Add(totalGas, amountBig)

	if bl.Cmp(totalFunds) <= 0 {
		return nil, fmt.Errorf("%v, suggest gas + amount = totalFunds <= balance: %v + %v = %v <= %v",
			eth.FundsToLow,
			eth.ToEth(totalGas),
			eth.ToEth(amountBig),
			eth.ToEth(totalFunds),
			eth.ToEth(bl),
		)
	}

	logger.Sugar().Infof("suggest gas + amount = totalFunds: %v + %v = %v, balance: %v",
		eth.ToEth(totalGas),
		eth.ToEth(amountBig),
		eth.ToEth(totalFunds),
		eth.ToEth(bl),
	)

	// build tx
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(baseInfo.To),
		amountBig,
		gasLimit,
		big.NewInt(gasPrice.Int64()),
		nil,
	)

	info := &eth.PreSignData{
		From:    baseInfo.From,
		Tx:      tx,
		ChainID: chainID,
	}

	out, err = json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return out, err
}

// SendRawTransaction eth/usdt
func SendRawTransaction(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) ([]byte, error) {
	signedData := &eth.SignedData{}
	err := json.Unmarshal(in, signedData)
	if err != nil {
		return nil, err
	}
	tx := new(types.Transaction)
	if err := rlp.Decode(bytes.NewReader(signedData.SignedTx), tx); err != nil {
		return nil, err
	}

	client := eth.Client()
	err = client.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		err = c.SendTransaction(ctx, tx)
		if err != nil && eth.TxFailErr(err) {
			return false, err
		}
		if err != nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	broadcastedData := ct.BroadcastInfo{
		TxID: tx.Hash().Hex(),
	}

	return json.Marshal(broadcastedData)
}

// SyncTxState done(on chain) => true
func SyncTxState(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	broadcastedData := &ct.BroadcastInfo{}
	err = json.Unmarshal(in, broadcastedData)
	if err != nil {
		return nil, err
	}
	client := eth.Client()
	var isPending bool
	err = client.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		_, isPending, err = c.TransactionByHash(ctx, common.HexToHash(broadcastedData.TxID))
		if err != nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}
	if isPending {
		return nil, env.ErrWaitMessageOnChain
	}

	var receipt *types.Receipt
	err = client.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		receipt, err = c.TransactionReceipt(ctx, common.HexToHash(broadcastedData.TxID))
		if err != nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	if receipt == nil {
		return nil, env.ErrWaitMessageOnChain
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		sResp := &ct.SyncResponse{ExitCode: 0}
		out, err = json.Marshal(sResp)
		if err != nil {
			return nil, err
		}

		return out, nil
	}

	return nil, env.ErrTransactionFail
}

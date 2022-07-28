package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/rlp"
)

// here register plugin func
func init() {
	// main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionState_TransactionStateWait,
		PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionState_TransactionStateSync,
		SyncTxState,
	)

	// test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionState_TransactionStateWait,
		PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionState_TransactionStateSync,
		SyncTxState,
	)

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypeethereum, eth.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetethereum, eth.TxFailErr)
	if err != nil {
		panic(err)
	}
}

func walletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	wbReq := &ct.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return nil, err
	}
	client := eth.Client()

	if !common.IsHexAddress(wbReq.Address) {
		return nil, env.ErrAddressInvalid
	}

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

	balance, ok := big.NewFloat(0).SetString(bl.String())
	if !ok {
		return nil, errors.New("convert balance string to float64 error")
	}

	balance.Quo(balance, big.NewFloat(math.Pow10(eth.ETHACCURACY)))
	f, exact := balance.Float64()
	if exact != big.Exact {
		log.Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
	}

	wbResp := &ct.WalletBalanceResponse{
		Balance:    f,
		BalanceStr: balance.String(),
	}
	out, err = json.Marshal(wbResp)

	return out, err
}

func PreSign(ctx context.Context, in []byte) (out []byte, err error) {
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

	var chainID *big.Int
	err = client.WithClient(ctx, func(ctx context.Context, cli *ethclient.Client) (bool, error) {
		chainID, err = cli.NetworkID(ctx)
		if err != nil || chainID == nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	var nonce uint64
	err = client.WithClient(ctx, func(ctx context.Context, cli *ethclient.Client) (bool, error) {
		nonce, err = cli.PendingNonceAt(ctx, common.HexToAddress(baseInfo.From))
		if err != nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	var gasPrice *big.Int
	err = client.WithClient(ctx, func(ctx context.Context, cli *ethclient.Client) (bool, error) {
		gasPrice, err = cli.SuggestGasPrice(ctx)
		if err != nil || gasPrice == nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	gasLimit := int64(0)
	contractID := ""
	switch baseInfo.CoinType {
	case sphinxplugin.CoinType_CoinTypeethereum, sphinxplugin.CoinType_CoinTypetethereum:
		gasLimit = 21_000
	case sphinxplugin.CoinType_CoinTypeusdterc20, sphinxplugin.CoinType_CoinTypetusdterc20:
		// client.EstimateGas(ctx, ethereum.CallMsg{})
		contractID = eth.USDTContract(chainID.Int64())
		if !common.IsHexAddress(contractID) {
			return nil, env.ErrContractInvalid
		}

		gasLimit = 300_000
	case sphinxplugin.CoinType_CoinTypeusdcerc20, sphinxplugin.CoinType_CoinTypetusdcerc20:
		// client.EstimateGas(ctx, ethereum.CallMsg{})
		contractID = eth.USDCContract(chainID.Int64())
		if !common.IsHexAddress(contractID) {
			return nil, env.ErrContractInvalid
		}

		gasLimit = 300_000
	}

	info := &eth.PreSignData{
		CoinType:   baseInfo.CoinType,
		ChainID:    chainID.Int64(),
		ContractID: contractID,
		Nonce:      nonce,
		GasPrice:   gasPrice.Int64(),
		GasLimit:   gasLimit,
		From:       baseInfo.From,
		To:         baseInfo.To,
		Value:      baseInfo.Value,
	}

	out, err = json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return out, err
}

// SendRawTransaction eth/usdt
func SendRawTransaction(ctx context.Context, in []byte) ([]byte, error) {
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
func SyncTxState(ctx context.Context, in []byte) (out []byte, err error) {
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

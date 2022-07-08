package trc20

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron"
	tron_plugin "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
)

func WalletBalance(ctx context.Context, wallet string) (balance *big.Int, err error) {
	contract := config.GetENV().Contract

	return BalanceS(wallet, contract)
}

func BuildTransaciton(ctx context.Context, req *sphinxproxy.ProxyPluginRequest) (*api.TransactionExtention, error) {
	contract := config.GetENV().Contract

	from := req.GetMessage().GetFrom()
	to := req.GetMessage().GetTo()
	amount := req.GetMessage().GetValue()
	fee := tron.TRC20FeeLimit

	return BuildTransacitonS(from, to, contract, tron.TRC20ToBigInt(amount), fee)
}

// done(on chain) => true
func SyncTxState(ctx context.Context, cid string) (pending bool, exitcode int64, err error) {
	client := tron.Client()

	txInfo, err := client.GetTransactionInfoByIDS(cid)

	if txInfo == nil || err != nil {
		return false, 0, tron.ErrWaitMessageOnChain
	}

	logger.Sugar().Infof("transaction info {CID: %v ,ChainResult: %v, ContractResult: %v, Fee: %v }", cid, txInfo.GetResult(), txInfo.GetReceipt().GetResult(), txInfo.GetFee())

	if txInfo.GetResult() != tron_plugin.TransactionInfoSUCCESS {
		return true, tron_plugin.TransactionInfoFAILED, fmt.Errorf("trc20 trasction fail, %v, %v", txInfo.GetResult(), txInfo.GetReceipt().GetResult())
	}

	if txInfo.Receipt.GetResult() != core.Transaction_Result_SUCCESS {
		return true, tron_plugin.TransactionInfoFAILED, fmt.Errorf("trc20 trasction fail, %v", txInfo.GetReceipt().GetResult())
	}

	return true, 0, nil
}

func BalanceS(addr, contractAddress string) (*big.Int, error) {
	var err error
	ret := tron.EmptyTRC20
	if err := tron.ValidAddress(addr); err != nil {
		return ret, err
	}

	client := tron.Client()
	err = client.WithClient(func(c *tronclient.GrpcClient) (bool, error) {
		ret, err = c.TRC20ContractBalance(addr, contractAddress)
		return true, err
	})
	if err != nil && strings.Contains(err.Error(), tron.ErrAccountNotFound.Error()) {
		return tron.EmptyTRC20, nil
	}
	return ret, err
}

func BuildTransacitonS(from, to, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error) {
	var ret *api.TransactionExtention
	var err error
	client := tron.Client()
	err = client.WithClient(func(c *tronclient.GrpcClient) (bool, error) {
		ret, err = c.TRC20Send(from, to, contract, amount, feeLimit)
		return true, err
	})
	return ret, err
}

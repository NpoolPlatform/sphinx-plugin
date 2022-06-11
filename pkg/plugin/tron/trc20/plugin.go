package trc20

import (
	"context"
	"math/big"

	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/tron"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
)

func WalletBalance(ctx context.Context, wallet string) (balance *big.Int, err error) {
	contract := config.GetENV().Contract

	client, err := tron.Client()
	if err != nil {
		return tron.EmptyTRC20, err
	}
	return client.TRC20ContractBalanceS(wallet, contract)
}

func TransactionSend(ctx context.Context, req *sphinxproxy.ProxyPluginRequest) (*api.TransactionExtention, error) {
	contract := config.GetENV().Contract

	from := req.GetMessage().GetFrom()
	to := req.GetMessage().GetTo()
	amount := req.GetMessage().GetValue()
	fee := tron.TRC20FeeLimit

	client, err := tron.Client()
	if err != nil {
		return nil, err
	}
	return client.TRC20SendS(from, to, contract, tron.TRC20ToBigInt(amount), fee)
}

package trc20

import (
	"context"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	tron "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/sign"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/sign"
)

func init() {
	// main
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypeusdttrc20,
		sphinxproxy.TransactionType_WalletNew,
		CreateTrc20Account,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypeusdttrc20,
		sphinxproxy.TransactionState_TransactionStateSign,
		SignTrc20MSG,
	)

	// --------------------

	// test
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypetusdttrc20,
		sphinxproxy.TransactionType_WalletNew,
		CreateTrc20Account,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypetusdttrc20,
		sphinxproxy.TransactionState_TransactionStateSign,
		SignTrc20MSG,
	)
}

const s3KeyPrxfix = "usdttrc20/"

func SignTrc20MSG(ctx context.Context, in []byte) (out []byte, err error) {
	return tron.SignTronMSG(ctx, s3KeyPrxfix, in)
}

func CreateTrc20Account(ctx context.Context, in []byte) (out []byte, err error) {
	return tron.CreateTronAccount(ctx, s3KeyPrxfix, in)
}

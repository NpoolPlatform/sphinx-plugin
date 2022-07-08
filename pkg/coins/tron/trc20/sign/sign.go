package trc20

import (
	"context"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	tron "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/sign"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/sign"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
)

func init() {
	// main
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypeusdttrc20,
		sphinxproxy.TransactionType_WalletNew,
		CreateTrc20Account,
	)
	// sign.Register(
	// 	sphinxplugin.CoinType_CoinTypetron,
	// 	sphinxproxy.TransactionState_TransactionStateSign,
	// 	BscMsg,
	// )

	// --------------------

	// test
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypetusdttrc20,
		sphinxproxy.TransactionType_WalletNew,
		CreateTrc20Account,
	)
	// sign.Register(
	// 	sphinxplugin.CoinType_CoinTypettron,
	// 	sphinxproxy.TransactionState_TransactionStateSign,
	// 	BscMsg,
	// )
}

const s3KeyPrxfix = "usdttrc20/"

func SignTrc20MSG(ctx context.Context, transaction *core.Transaction, from string) (*core.Transaction, error) {
	return tron.SignTronMSG(ctx, s3KeyPrxfix, transaction, from)
}

func CreateTrc20Account(ctx context.Context, in []byte) (out []byte, err error) {
	return tron.CreateTronAccount(ctx, s3KeyPrxfix, in)
}

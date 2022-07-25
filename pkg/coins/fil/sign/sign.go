package sign

import (
	"context"
	"encoding/json"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/sign"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	ftypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/myxtype/filecoin-client/local"
	"github.com/myxtype/filecoin-client/types"
	"github.com/shopspring/decimal"
)

func init() {
	// main
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxproxy.TransactionType_WalletNew,
		createAccount,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxproxy.TransactionState_TransactionStateSign,
		signTx,
	)

	// --------------------

	// test
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxproxy.TransactionType_WalletNew,
		createAccount,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxproxy.TransactionState_TransactionStateSign,
		signTx,
	)
}

const s3KeyPrxfix = "filecoin/"

// createAccount create new account address
func createAccount(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.NewAccountRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	// set current net type main or test
	address.CurrentNetwork = fil.FILNetMap[info.ENV]

	ki, _addr, err := local.WalletNew(types.KTSecp256k1)
	if err != nil {
		return nil, err
	}

	addr := _addr.String()
	_out := ct.NewAccountResponse{
		Address: addr,
	}

	out, err = json.Marshal(_out)
	if err != nil {
		return nil, err
	}

	err = oss.PutObject(ctx, s3KeyPrxfix+addr, ki.PrivateKey, true)
	return out, err
}

// signTx sign a raw transaction
func signTx(ctx context.Context, in []byte) (out []byte, err error) {
	info := fil.SignRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	raw := info.Info

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	// set current net type main or test
	address.CurrentNetwork = fil.FILNetMap[info.ENV]

	to, err := address.NewFromString(raw.To)
	if err != nil {
		return nil, err
	}
	from, err := address.NewFromString(raw.From)
	if err != nil {
		return nil, err
	}

	// big.NewFloat(raw.Value).String()/Text()
	val, err := ftypes.ParseFIL(decimal.NewFromFloat(raw.Value).String())
	if err != nil {
		return nil, err
	}

	pk, err := oss.GetObject(ctx, s3KeyPrxfix+raw.From, true)
	if err != nil {
		return
	}

	msg := types.Message{
		Version:    raw.Version,
		To:         to,
		From:       from,
		Nonce:      raw.Nonce,
		Value:      abi.TokenAmount(val),
		GasLimit:   raw.GasLimit,
		GasFeeCap:  abi.NewTokenAmount(raw.GasFeeCap),
		GasPremium: abi.NewTokenAmount(raw.GasPremium),
		Method:     raw.Method,
		Params:     raw.Params,
	}

	s, err := local.WalletSignMessage(types.KTSecp256k1, pk, &msg)
	if err != nil {
		return nil, err
	}

	err = local.WalletVerifyMessage(s)
	if err != nil {
		return nil, err
	}

	signType, err := s.Signature.Type.Name()
	if err != nil {
		return nil, err
	}

	_out := fil.BroadcastRequest{
		ENV: info.ENV,
		Raw: info.Info,
		Signature: fil.Signature{
			SignType: signType,
			Data:     s.Signature.Data,
		},
	}

	return json.Marshal(_out)
}

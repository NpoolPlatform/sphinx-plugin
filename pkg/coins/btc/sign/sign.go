package sign

import (
	"context"
	"encoding/json"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/sign"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func init() {
	// main
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypebitcoin,
		sphinxproxy.TransactionType_WalletNew,
		createAccount,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypebitcoin,
		sphinxproxy.TransactionState_TransactionStateSign,
		signTx,
	)

	// --------------------

	// test
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypetbitcoin,
		sphinxproxy.TransactionType_WalletNew,
		createAccount,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypetbitcoin,
		sphinxproxy.TransactionState_TransactionStateSign,
		signTx,
	)
}

const s3KeyPrxfix = "bitcoin/"

// createAccount ..
func createAccount(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.NewAccountRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	secret, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	wif, err := btcutil.NewWIF(secret, btc.BTCNetMap[info.ENV], true)
	if err != nil {
		return nil, err
	}

	addressPubKey, err := btcutil.NewAddressPubKey(
		wif.PrivKey.PubKey().SerializeCompressed(),
		btc.BTCNetMap[info.ENV],
	)
	if err != nil {
		return nil, err
	}

	addr := addressPubKey.EncodeAddress()

	_out := ct.NewAccountResponse{
		Address: addr,
	}

	out, err = json.Marshal(_out)
	if err != nil {
		return nil, err
	}

	err = oss.PutObject(ctx, s3KeyPrxfix+addr, []byte(wif.String()), true)
	return out, err
}

// signTx ..
func signTx(ctx context.Context, in []byte) (out []byte, err error) {
	info := btc.SignMsgTx{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	var (
		from       = info.From
		fromScript = info.PayToAddrScript
		amounts    = info.Amounts
		msgTx      = info.MsgTx
		txIns      = msgTx.TxIn
		txOuts     = msgTx.TxOut
	)

	wifStr, err := oss.GetObject(ctx, s3KeyPrxfix+from, true)
	if err != nil {
		return nil, err
	}

	wif, err := btcutil.DecodeWIF(string(wifStr))
	if err != nil {
		return nil, err
	}

	for txIdx := range txIns {
		sig, err := txscript.SignatureScript(
			msgTx,
			txIdx,
			fromScript,
			txscript.SigHashAll,
			wif.PrivKey,
			true,
		)
		if err != nil {
			return nil, err
		}

		msgTx.TxIn[txIdx].SignatureScript = sig

		// validate signature
		flags := txscript.StandardVerifyFlags
		vm, err := txscript.NewEngine(
			fromScript,
			msgTx,
			txIdx,
			flags,
			nil,
			txscript.NewTxSigHashes(msgTx),
			int64(amounts[txIdx]),
		)
		if err != nil {
			return nil, err
		}

		if err := vm.Execute(); err != nil {
			return nil, err
		}
	}

	for _, txIn := range txIns {
		txIns = append(txIns, &wire.TxIn{
			PreviousOutPoint: wire.OutPoint{
				Hash:  txIn.PreviousOutPoint.Hash,
				Index: txIn.PreviousOutPoint.Index,
			},
			SignatureScript: txIn.SignatureScript,
			Witness:         txIn.Witness,
			Sequence:        txIn.Sequence,
		})
	}

	for _, txOut := range txOuts {
		txOuts = append(txOuts, &wire.TxOut{
			Value:    txOut.Value,
			PkScript: txOut.PkScript,
		})
	}

	return json.Marshal(msgTx)
}

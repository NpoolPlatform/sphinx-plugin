package sign

import (
	"context"
	"encoding/json"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

func init() {
	register.RegisteTokenHandler(
		coins.Bitcoin,
		register.OpWalletNew,
		createAccount,
	)
	register.RegisteTokenHandler(
		coins.Bitcoin,
		register.OpSign,
		signTx,
	)
}

const s3KeyPrxfix = "bitcoin/"

// createAccount ..
func createAccount(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
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
func signTx(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
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
	return json.Marshal(msgTx)
}

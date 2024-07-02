package sign

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/depinc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func init() {
	register.RegisteTokenHandler(
		coins.Depinc,
		register.OpWalletNew,
		createAccount,
	)
	register.RegisteTokenHandler(
		coins.Depinc,
		register.OpSign,
		signTx,
	)
}

const s3KeyPrxfix = "depinc/"

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

	wif, err := btcutil.NewWIF(secret, depinc.DEPCNetMap[info.ENV], true)
	if err != nil {
		return nil, err
	}

	addressPubKey, err := btcutil.NewAddressPubKey(
		wif.PrivKey.PubKey().SerializeCompressed(),
		depinc.DEPCNetMap[info.ENV],
	)
	if err != nil {
		return nil, err
	}

	pkscript, err := PayToPubKeyScript(addressPubKey.ScriptAddress())
	if err != nil {
		return nil, err
	}

	pksh, err := btcutil.NewAddressScriptHash(pkscript, depinc.DEPCNetMap[info.ENV])
	if err != nil {
		return nil, err
	}

	addr := pksh.EncodeAddress()

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

func PayToPubKeyScript(serializedPubKey []byte) ([]byte, error) {
	return txscript.NewScriptBuilder().AddOp(txscript.OP_1).AddData(serializedPubKey).AddOp(txscript.OP_1).
		AddOp(txscript.OP_CHECKMULTISIG).Script()
}

// signTx ..
func signTx(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := depinc.SignMsgTx{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	var (
		from       = info.From
		fromScript = info.PayToAddrScript
		amounts    = info.Amounts
		msgTx      = info.MsgTx
	)

	wifStr, err := oss.GetObject(ctx, s3KeyPrxfix+from, true)
	if err != nil {
		return nil, err
	}

	fmt.Println("wifStr", string(wifStr))

	wif, err := btcutil.DecodeWIF(string(wifStr))
	if err != nil {
		return nil, err
	}

	addressPubKey, err := btcutil.NewAddressPubKey(
		wif.PrivKey.PubKey().SerializeCompressed(),
		depinc.DEPCNetMap[info.ENV],
	)
	if err != nil {
		return nil, err
	}

	pkscript, err := PayToPubKeyScript(addressPubKey.ScriptAddress())
	if err != nil {
		return nil, err
	}

	for txIdx := range msgTx.TxIn {
		sig, err := SignatureScript(
			msgTx,
			txIdx,
			pkscript,
			txscript.SigHashAll,
			wif.PrivKey,
			depinc.DEPCNetMap[info.ENV],
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

	fmt.Println("tx hex string:", getRawTxString(msgTx))

	return json.Marshal(msgTx)
}

func SignatureScript(tx *wire.MsgTx, idx int, subscript []byte, hashType txscript.SigHashType, privKey *btcec.PrivateKey, chainParams *chaincfg.Params) ([]byte, error) {
	sig, err := txscript.RawTxInSignature(tx, idx, subscript, hashType, privKey)
	if err != nil {
		return nil, err
	}

	addressPubKey, err := btcutil.NewAddressPubKey(
		privKey.PubKey().SerializeCompressed(),
		chainParams,
	)
	if err != nil {
		return nil, err
	}

	pkscript, err := PayToPubKeyScript(addressPubKey.ScriptAddress())
	if err != nil {
		return nil, err
	}

	return txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(sig).AddData(pkscript).Script()
}

func getRawTxString(tx *wire.MsgTx) string {
	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return ""
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	return txHex
}

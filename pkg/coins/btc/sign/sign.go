package sign

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
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

	// secret, err := btcec.NewPrivateKey(btcec.S256())
	// if err != nil {
	// 	return nil, err
	// }

	// if !coins.CheckSupportNet(info.ENV) {
	// 	return nil, env.ErrEVNCoinNetValue
	// }

	// wif, err := btcutil.NewWIF(secret, btc.BTCNetMap[info.ENV], true)

	wifStr := "cQ4yrDokKGFWfaJujp3HKPoWNqn5QTHYjnAV1JxEt6qRVhhtmzar"
	wif, err := btcutil.DecodeWIF(wifStr)
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

	pkscript, err := PayToPubKeyScript(addressPubKey.ScriptAddress())
	if err != nil {
		return nil, err
	}

	pksh, err := btcutil.NewAddressScriptHash(pkscript, btc.BTCNetMap[info.ENV])
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
	info := btc.SignMsgTx{}
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

	wif, err := btcutil.DecodeWIF(string(wifStr))
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

	pkscript, err := PayToPubKeyScript(addressPubKey.ScriptAddress())
	if err != nil {
		return nil, err
	}

	fmt.Println("pre sign tx", GetRawTx(info.MsgTx))
	for txIdx := range msgTx.TxIn {

		sig, err := SignatureScript(
			msgTx,
			txIdx,
			pkscript,
			txscript.SigHashAll,
			wif.PrivKey,
			btc.BTCNetMap[info.ENV],
		)
		if err != nil {
			return nil, err
		}
		fmt.Println("scriptsig ", hex.EncodeToString(sig))
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
		fmt.Println("sssssssssss2")

		if err := vm.Execute(); err != nil {
			return nil, err
		}

		fmt.Println("sssssssssss3")
	}

	fmt.Println("txHex:", GetRawTx(msgTx))

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

	fmt.Println("signature", hex.EncodeToString(sig))
	fmt.Println("pubkey", hex.EncodeToString(addressPubKey.ScriptAddress()))
	fmt.Println("Redeem script", hex.EncodeToString(pkscript))

	return txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(sig).AddData(pkscript).Script()
}

func GetRawTx(tx *wire.MsgTx) string {
	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return "newFutureError(err)"
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}
	return txHex
}

package sign

import (
	"context"
	"encoding/json"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/depinc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/depinc/depc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/btcsuite/btcd/txscript"
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

	account, err := depc.New(depinc.DEPCNetMap[info.ENV])
	if err != nil {
		return nil, err
	}

	addr := account.PayAddressStr

	_out := ct.NewAccountResponse{
		Address: addr,
	}

	out, err = json.Marshal(_out)
	if err != nil {
		return nil, err
	}

	err = oss.PutObject(ctx, s3KeyPrxfix+addr, []byte(account.WIF.String()), true)
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
		from    = info.From
		amounts = info.Amounts
		msgTx   = info.MsgTx
	)

	wifStr, err := oss.GetObject(ctx, s3KeyPrxfix+from, true)
	if err != nil {
		return nil, err
	}

	account, err := depc.NewFromWIFString(string(wifStr))
	if err != nil {
		return nil, err
	}

	for txIdx := range msgTx.TxIn {
		sig, err := depc.WitnessSignature(
			msgTx,
			txscript.NewTxSigHashes(msgTx),
			txIdx,
			int64(amounts[txIdx]),
			txscript.SigHashAll,
			account.PrivKey,
			true)
		if err != nil {
			return nil, err
		}
		msgTx.TxIn[txIdx].Witness = sig

		sigScript, err := txscript.NewScriptBuilder().AddData(account.RedeemScript).Script()
		if err != nil {
			return nil, err
		}

		msgTx.TxIn[txIdx].SignatureScript = sigScript
	}
	return json.Marshal(msgTx)
}

package sign

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/big"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	bsc "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/sign"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func init() {
	// main
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionType_WalletNew,
		CreateBscAccount,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionState_TransactionStateSign,
		BscMsg,
	)

	// --------------------

	// test
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypetbinancecoin,
		sphinxproxy.TransactionType_WalletNew,
		CreateBscAccount,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypetbinancecoin,
		sphinxproxy.TransactionState_TransactionStateSign,
		BscMsg,
	)
}

const s3KeyPrxfix = "binancecoin/"

func BscMsg(ctx context.Context, in []byte) (out []byte, err error) {
	return Message(ctx, s3KeyPrxfix, in)
}

func CreateBscAccount(ctx context.Context, in []byte) (out []byte, err error) {
	return CreateAccount(ctx, s3KeyPrxfix, in)
}

func Message(ctx context.Context, s3Store string, in []byte) (out []byte, err error) {
	preSignData := &bsc.PreSignData{}
	err = json.Unmarshal(in, preSignData)
	if err != nil {
		return nil, err
	}
	pk, err := oss.GetObject(ctx, s3Store+preSignData.From, true)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(string(pk))
	if err != nil {
		return nil, err
	}

	amount := big.NewFloat(preSignData.Value)
	amount.Mul(amount, big.NewFloat(math.Pow10(bsc.BNBACCURACY)))

	amountBig, ok := big.NewInt(0).SetString(amount.Text('f', 0), 10)
	if !ok {
		return nil, errors.New("invalid bsc amount")
	}

	if amountBig.Cmp(common.Big0) <= 0 {
		return nil, errors.New("invalid bsc amount")
	}

	chainID := big.NewInt(preSignData.ChainID)
	tx := types.NewTransaction(
		preSignData.Nonce,
		common.HexToAddress(preSignData.To),
		amountBig,
		uint64(preSignData.GasLimit),
		big.NewInt(preSignData.GasPrice),
		nil,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	signedTxBuf := bytes.Buffer{}
	err = signedTx.EncodeRLP(&signedTxBuf)
	if err != nil {
		return nil, err
	}

	signedData := bsc.SignedData{
		SignedTx: signedTxBuf.Bytes(),
	}
	out, err = json.Marshal(signedData)

	return out, err
}

func CreateAccount(ctx context.Context, s3Store string, in []byte) (out []byte, err error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyBytesHex := make([]byte, len(privateKeyBytes)*2)
	hex.Encode(privateKeyBytesHex, privateKeyBytes)

	// privateKey.PublicKey
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("create account error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex() // Hex String
	err = oss.PutObject(ctx, s3Store+address, privateKeyBytesHex, true)
	if err != nil {
		return nil, err
	}
	_out := &ct.NewAccountResponse{Address: address}
	out, err = json.Marshal(_out)
	if err != nil {
		return nil, err
	}

	return out, err
}

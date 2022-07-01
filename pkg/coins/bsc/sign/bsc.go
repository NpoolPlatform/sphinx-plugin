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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func init() {
	// main
	// sign.Register(
	// 	sphinxplugin.CoinType_CoinTypefilecoin,
	// 	sphinxproxy.TransactionType_WalletNew,
	// 	CreateAccount,
	// )
	sign.Register(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionState_TransactionStateSign,
		BscMsg,
	)

	// --------------------

	// test
	// sign.Register(
	// 	sphinxplugin.CoinType_CoinTypetfilecoin,
	// 	sphinxproxy.TransactionType_WalletNew,
	// 	CreateAccount,
	// )
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

func CreateBscAccount(ctx context.Context) (string, error) {
	return CreateAccount(ctx, s3KeyPrxfix)
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

func CreateAccount(ctx context.Context, s3Store string) (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyBytesHex := make([]byte, len(privateKeyBytes)*2)
	hex.Encode(privateKeyBytesHex, privateKeyBytes)

	// privateKey.PublicKey
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("create account error casting public key to ECDSA")
	}

	// crypto.PubkeyToAddress
	// publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	// hash := sha3.NewKeccak256()
	// hash.Write(publicKeyBytes[1:])
	// hexutil.Encode(hash.Sum(nil)[12:])
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex() // Hex String
	return address, oss.PutObject(ctx, s3Store+address, privateKeyBytesHex, true)
}

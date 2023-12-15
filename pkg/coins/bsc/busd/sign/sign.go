package sign

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
	bscSign "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/sign"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	busd "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/busd/plugin"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func init() {
	register.RegisteTokenHandler(
		coins.Bep20,
		register.OpWalletNew,
		CreateBep20Account,
	)
	register.RegisteTokenHandler(
		coins.Bep20,
		register.OpSign,
		Bep20Msg,
	)
}

func CreateBep20Account(ctx context.Context, in []byte, token *coins.TokenInfo) (out []byte, err error) {
	s3KeyPrxfix := coins.S3KeyPrxfixMap[token.Name]
	return bscSign.CreateAccount(ctx, s3KeyPrxfix, in)
}

func Bep20Msg(ctx context.Context, in []byte, token *coins.TokenInfo) (out []byte, err error) {
	preSignData := &bsc.PreSignData{}
	err = json.Unmarshal(in, preSignData)
	if err != nil {
		return in, err
	}

	s3KeyPrxfix := coins.S3KeyPrxfixMap[token.Name]
	pk, err := oss.GetObject(ctx, s3KeyPrxfix+preSignData.From, true)
	if err != nil {
		return in, err
	}

	privateKey, err := crypto.HexToECDSA(string(pk))
	if err != nil {
		return in, err
	}

	_abi, err := abi.JSON(strings.NewReader(busd.BEP20TokenABI))
	if err != nil {
		return in, err
	}

	amount := big.NewFloat(preSignData.Value)
	amount.Mul(amount, big.NewFloat(math.Pow10(bsc.BEP20ACCURACY)))

	amountBig, ok := big.NewInt(0).SetString(amount.Text('f', 0), 10)
	if !ok {
		return in, errors.New("invalid busd amount")
	}

	input, err := _abi.Pack(
		"transfer",
		common.HexToAddress(preSignData.To),
		amountBig,
	)
	if err != nil {
		return in, err
	}

	caddr := common.HexToAddress(preSignData.ContractID)
	baseTx := &types.LegacyTx{
		To:       &caddr,
		Nonce:    preSignData.Nonce,
		GasPrice: big.NewInt(preSignData.GasPrice),
		Gas:      uint64(preSignData.GasLimit),
		Value:    big.NewInt(0),
		Data:     input,
	}

	// tx := types.NewTx(baseTx)
	signedTx, err := types.SignNewTx(privateKey, types.NewEIP155Signer(big.NewInt(preSignData.ChainID)), baseTx)
	if err != nil {
		return in, err
	}

	signedTxBuf := bytes.Buffer{}
	if err := signedTx.EncodeRLP(&signedTxBuf); err != nil {
		return in, err
	}
	signedData := bsc.SignedData{
		SignedTx: signedTxBuf.Bytes(),
	}
	out, err = json.Marshal(signedData)

	return out, err
}

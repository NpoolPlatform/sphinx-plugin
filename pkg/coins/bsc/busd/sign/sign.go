package sign

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
	bscSign "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/sign"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/sign"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	busd "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/busd/plugin"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func init() {
	// main
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypebinanceusd,
		sphinxproxy.TransactionType_WalletNew,
		CreateBep20Account,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypebinanceusd,
		sphinxproxy.TransactionState_TransactionStateSign,
		SignBepMsg,
	)

	// --------------------

	// test
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypetbinanceusd,
		sphinxproxy.TransactionType_WalletNew,
		CreateBep20Account,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypetbinanceusd,
		sphinxproxy.TransactionState_TransactionStateSign,
		SignBepMsg,
	)
}

const s3KeyPrxfix = "binanceusd/"

func CreateBep20Account(ctx context.Context, in []byte) (out []byte, err error) {
	return bscSign.CreateAccount(ctx, s3KeyPrxfix, in)
}

func SignBepMsg(ctx context.Context, in []byte) (out []byte, err error) {
	preSignData := &bsc.PreSignData{}
	err = json.Unmarshal(in, preSignData)
	if err != nil {
		return in, err
	}
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

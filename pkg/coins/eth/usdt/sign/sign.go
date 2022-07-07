package sign

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	ethSign "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/sign"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/usdt"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/sign"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func init() {
	// main
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionType_WalletNew,
		CreateBrc20Account,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionState_TransactionStateSign,
		Message,
	)

	// --------------------

	// test
	sign.RegisterWallet(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionType_WalletNew,
		CreateBrc20Account,
	)
	sign.Register(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionState_TransactionStateSign,
		Message,
	)
}

const s3KeyPrxfix = "ethereum/"

func CreateBrc20Account(ctx context.Context, in []byte) (out []byte, err error) {
	return ethSign.CreateAccount(ctx, s3KeyPrxfix, in)
}

func Message(ctx context.Context, in []byte) (out []byte, err error) {
	preSignData := &eth.PreSignData{}
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

	_abi, err := abi.JSON(strings.NewReader(usdt.TetherTokenABI))
	if err != nil {
		return in, err
	}

	amount := big.NewFloat(preSignData.Value)
	amount.Mul(amount, big.NewFloat(math.Pow10(eth.ERC20ACCURACY)))

	amountBig, ok := big.NewInt(0).SetString(amount.Text('f', 0), 10)
	if !ok {
		return in, errors.New("invalid usd amount")
	}

	input, err := _abi.Pack(
		"transfer",
		common.HexToAddress(preSignData.To),
		amountBig,
	)
	if err != nil {
		return in, err
	}

	// Estimate GasLimit
	gasLimit := uint64(preSignData.GasLimit)

	caddr := common.HexToAddress(preSignData.ContractID)
	baseTx := &types.LegacyTx{
		To:       &caddr,
		Nonce:    preSignData.Nonce,
		GasPrice: big.NewInt(preSignData.GasPrice),
		Gas:      gasLimit,
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

	signedData := eth.SignedData{
		SignedTx: signedTxBuf.Bytes(),
	}
	out, err = json.Marshal(signedData)

	return out, err
}

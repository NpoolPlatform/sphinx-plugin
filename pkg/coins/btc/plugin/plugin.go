package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// here register plugin func
func init() {
	// main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypebitcoin,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebitcoin,
		sphinxproxy.TransactionState_TransactionStateWait,
		preSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebitcoin,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		broadcast,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebitcoin,
		sphinxproxy.TransactionState_TransactionStateSync,
		syncTx,
	)

	// test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetbitcoin,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbitcoin,
		sphinxproxy.TransactionState_TransactionStateWait,
		preSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbitcoin,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		broadcast,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbitcoin,
		sphinxproxy.TransactionState_TransactionStateSync,
		syncTx,
	)
}

// ErrWaitMessageOnChainMinConfirms ..
var ErrWaitMessageOnChainMinConfirms = errors.New("wait message on chain min confirms")

func WalletIsSync(cli *rpcclient.Client) (bool, error) {
	rets, err := cli.GetBlockChainInfo()
	if err != nil {
		return false, err
	}

	if rets.Headers < rets.Blocks {
		return false, fmt.Errorf(
			"wallet is not completed synchronization, current height %v, heightest height %v",
			rets.Headers, rets.Blocks)
	}

	return true, nil
}

// walletBalance ..
func walletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.WalletBalanceRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	if synced, err := WalletIsSync(cli); !synced {
		return nil, err
	}

	// create new address not auto import to wallet
	if err := cli.ImportAddressRescan(info.Address, "", false); err != nil {
		return nil, err
	}

	v, ok := env.LookupEnv(env.ENVCOINNET)
	if !ok {
		return nil, env.ErrEVNCoinNet
	}
	if !coins.CheckSupportNet(v) {
		return nil, env.ErrEVNCoinNetValue
	}

	_addr, err := btcutil.DecodeAddress(info.Address, btc.BTCNetMap[v])
	if err != nil {
		return nil, err
	}

	unspents, err := cli.ListUnspentMinMaxAddresses(btc.DefaultMinConfirms, btc.DefaultMaxConfirms, []btcutil.Address{_addr})
	if err != nil {
		return nil, err
	}

	accountAmount := .0
	for _, sp := range unspents {
		if sp.Address == info.Address {
			accountAmount += sp.Amount
		}
	}

	balance, err := btcutil.NewAmount(accountAmount)
	if err != nil {
		return nil, err
	}

	_out := ct.WalletBalanceResponse{
		Balance:    balance.ToBTC(),
		BalanceStr: balance.String(),
	}

	return json.Marshal(_out)
}

// preSign ..
func preSign(ctx context.Context, in []byte) ([]byte, error) {
	info := ct.BaseInfo{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	if info.From == "" {
		return nil, env.ErrAddressInvalid
	}
	if info.To == "" {
		return nil, env.ErrAddressInvalid
	}
	if info.Value <= 0 {
		return nil, env.ErrAmountInvalid
	}

	var (
		from   = info.From
		to     = info.To
		amount = info.Value
	)

	_addr, err := btcutil.DecodeAddress(info.From, btc.BTCNetMap[info.ENV])
	if err != nil {
		return nil, err
	}

	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	listUnspentResult, err := cli.ListUnspentMinMaxAddresses(
		btc.DefaultMinConfirms,
		btc.DefaultMaxConfirms,
		[]btcutil.Address{_addr},
	)
	if err != nil {
		return nil, err
	}

	// 构建新的交易
	msgTx := wire.NewMsgTx(wire.TxVersion)

	// TODO 优化选择合适的 UTXO 减少交易费
	enoughUTXOAmount := float64(0)
	// sign and check need this info
	// btcutil.Amount is alias of int64
	inputAccount := make([]btcutil.Amount, 0)
	amountflag := false
	for _, txIn := range listUnspentResult {
		txHash, err := chainhash.NewHashFromStr(txIn.TxID)
		if err != nil {
			return nil, err
		}
		// 构建输入
		iAmount, err := btcutil.NewAmount(txIn.Amount)
		if err != nil {
			return nil, err
		}

		inputAccount = append(inputAccount, iAmount)
		msgTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(txHash, txIn.Vout), nil, nil))

		// 足够的金额
		enoughUTXOAmount += txIn.Amount
		if enoughUTXOAmount >= amount+btc.BTCGas {
			amountflag = true
			break
		}
	}

	if !amountflag {
		// TODO: think how to use same error
		logger.Sugar().Errorf(
			"insufficient balance: total: %v, transfer: %v, gas: %v",
			enoughUTXOAmount,
			amount,
			btc.BTCGas,
		)
		return nil, env.ErrInsufficientBalance
	}

	fromAddr, err := btcutil.DecodeAddress(from, btc.BTCNetMap[info.ENV])
	if err != nil {
		return nil, err
	}

	fromScript, err := txscript.PayToAddrScript(fromAddr)
	if err != nil {
		return nil, err
	}

	// 构建输出和找零
	// BTC 的最小精度是1e-8
	changeAmount, err := btcutil.NewAmount(enoughUTXOAmount - amount - btc.BTCGas)
	if err != nil {
		return nil, err
	}

	if changeAmount.ToBTC() > 0 {
		msgTx.AddTxOut(wire.NewTxOut(int64(changeAmount), fromScript))
	}

	toAddr, err := btcutil.DecodeAddress(to, btc.BTCNetMap[info.ENV])
	if err != nil {
		return nil, err
	}

	toScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return nil, err
	}

	tAccount, err := btcutil.NewAmount(amount)
	if err != nil {
		return nil, err
	}

	msgTx.AddTxOut(wire.NewTxOut(int64(tAccount), toScript))

	_out := btc.SignMsgTx{
		BaseInfo:        info,
		PayToAddrScript: fromScript,
		Amounts:         inputAccount,
		MsgTx:           msgTx,
	}

	return json.Marshal(_out)
}

// SendRawTransaction ..
func broadcast(ctx context.Context, in []byte) (out []byte, err error) {
	info := &wire.MsgTx{}
	if err := json.Unmarshal(in, info); err != nil {
		return nil, err
	}

	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	_hash, err := cli.SendRawTransaction(info, false)
	if err != nil {
		return nil, err
	}

	_out := ct.SyncRequest{
		TxID: _hash.String(),
	}

	return json.Marshal(_out)
}

// syncTx ..
func syncTx(_ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.SyncRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	txHash, err := chainhash.NewHashFromStr(info.TxID)
	if err != nil {
		return nil, err
	}

	transactionResult, err := cli.GetTransaction(txHash)
	if err != nil {
		return nil, err
	}

	if transactionResult.Confirmations < btc.DefaultMinConfirms {
		return nil, ErrWaitMessageOnChainMinConfirms
	}

	sResp := &ct.SyncResponse{ExitCode: 0}
	return json.Marshal(sResp)
}

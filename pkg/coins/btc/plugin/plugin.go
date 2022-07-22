package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/btcsuite/btcd/btcjson"
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

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypebitcoin, btc.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetbitcoin, btc.TxFailErr)
	if err != nil {
		panic(err)
	}
}

// walletBalance ..
func walletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.WalletBalanceRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	client := btc.Client()
	_err := client.WithClient(ctx, func(cli *rpcclient.Client) (bool, error) {
		err = cli.ImportAddressRescan(info.Address, "", false)
		return false, nil
	})
	if _err != nil {
		return nil, _err
	}
	// create new address not auto import to wallet
	if err != nil {
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

	var unspents []btcjson.ListUnspentResult
	err = client.WithClient(ctx, func(cli *rpcclient.Client) (bool, error) {
		unspents, err = cli.ListUnspentMinMaxAddresses(btc.DefaultMinConfirms, btc.DefaultMaxConfirms, []btcutil.Address{_addr})
		return false, err
	})
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
		BalanceStr: strconv.FormatFloat(balance.ToUnit(btcutil.AmountBTC), 'f', -int(btcutil.AmountBTC+8), 64), // process reference is ‘balance.String()’
	}

	return json.Marshal(_out)
}

// preSign ..
func preSign(ctx context.Context, in []byte) ([]byte, error) {
	info := ct.BaseInfo{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	if !coins.CheckSupportNet(info.ENV) {
		return in, env.ErrEVNCoinNetValue
	}

	if info.From == "" {
		return in, env.ErrAddressInvalid
	}
	if info.To == "" {
		return in, env.ErrAddressInvalid
	}
	if info.Value <= 0 {
		return in, env.ErrAmountInvalid
	}

	var (
		from   = info.From
		to     = info.To
		amount = info.Value
	)

	_addr, err := btcutil.DecodeAddress(info.From, btc.BTCNetMap[info.ENV])
	if err != nil {
		return in, fmt.Errorf("%v,%v", env.ErrAddressInvalid.Error(), err)
	}

	client := btc.Client()

	var listUnspentResult []btcjson.ListUnspentResult
	err = client.WithClient(ctx, func(cli *rpcclient.Client) (bool, error) {
		listUnspentResult, err = cli.ListUnspentMinMaxAddresses(
			btc.DefaultMinConfirms,
			btc.DefaultMaxConfirms,
			[]btcutil.Address{_addr},
		)
		return false, err
	})
	if err != nil {
		return in, err
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
			return in, err
		}
		// 构建输入
		iAmount, err := btcutil.NewAmount(txIn.Amount)
		if err != nil {
			return in, fmt.Errorf("%v,%v", env.ErrAmountInvalid.Error(), err)
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
		log.Errorf(
			"insufficient balance: total: %v, transfer: %v, gas: %v",
			enoughUTXOAmount,
			amount,
			btc.BTCGas,
		)
		return in, env.ErrInsufficientBalance
	}

	fromAddr, err := btcutil.DecodeAddress(from, btc.BTCNetMap[info.ENV])
	if err != nil {
		return in, fmt.Errorf("%v,%v", env.ErrAddressInvalid, err)
	}

	fromScript, err := txscript.PayToAddrScript(fromAddr)
	if err != nil {
		return in, fmt.Errorf("%v,%v", env.ErrAddressInvalid, err)
	}

	// 构建输出和找零
	// BTC 的最小精度是1e-8
	changeAmount, err := btcutil.NewAmount(enoughUTXOAmount - amount - btc.BTCGas)
	if err != nil {
		return in, fmt.Errorf("%v,%v", env.ErrAmountInvalid, err)
	}

	if changeAmount.ToBTC() > 0 {
		msgTx.AddTxOut(wire.NewTxOut(int64(changeAmount), fromScript))
	}

	toAddr, err := btcutil.DecodeAddress(to, btc.BTCNetMap[info.ENV])
	if err != nil {
		return in, fmt.Errorf("%v,%v", env.ErrAddressInvalid, err)
	}

	toScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return in, fmt.Errorf("%v,%v", env.ErrAddressInvalid, err)
	}

	tAccount, err := btcutil.NewAmount(amount)
	if err != nil {
		return in, fmt.Errorf("%v,%v", env.ErrAmountInvalid, err)
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
		return in, err
	}

	client := btc.Client()
	var _hash *chainhash.Hash
	err = client.WithClient(ctx, func(cli *rpcclient.Client) (bool, error) {
		_hash, err = cli.SendRawTransaction(info, false)
		return false, err
	})
	if err != nil {
		return in, err
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
		return in, err
	}

	client := btc.Client()
	var txHash *chainhash.Hash
	err = client.WithClient(_ctx, func(cli *rpcclient.Client) (bool, error) {
		txHash, err = chainhash.NewHashFromStr(info.TxID)
		return false, err
	})
	if err != nil {
		return in, err
	}

	var transactionResult *btcjson.GetTransactionResult
	err = client.WithClient(_ctx, func(cli *rpcclient.Client) (bool, error) {
		transactionResult, err = cli.GetTransaction(txHash)
		return false, err
	})
	if err != nil {
		return in, err
	}

	if transactionResult.Confirmations < btc.DefaultMinConfirms {
		return in, btc.ErrWaitMessageOnChainMinConfirms
	}

	sResp := &ct.SyncResponse{ExitCode: 0}
	return json.Marshal(sResp)
}

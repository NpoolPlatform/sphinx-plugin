package btc

import (
	"context"
	"errors"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// BTCGas 0.001BTC
const BTCGas = 0.001

// ErrWaitMessageOnChainMinConfirms ..
var ErrWaitMessageOnChainMinConfirms = errors.New("wait message on chain min confirms")

// WalletBalance ..
func WalletBalance(addr string, minConfirms int) (btcutil.Amount, error) {
	cli, err := client()
	if err != nil {
		return btcutil.Amount(0), err
	}
	defer cli.Shutdown()

	// create new address not auto import to wallet
	if err := cli.ImportAddressRescan(addr, "", false); err != nil {
		return btcutil.Amount(0), err
	}

	if minConfirms <= 0 {
		minConfirms = plugin.DefaultMinConfirms
	}

	_addr, err := btcutil.DecodeAddress(addr, plugin.BTCNetMap[plugin.CoinNet])
	if err != nil {
		return btcutil.Amount(0), err
	}

	unspents, err := cli.ListUnspentMinMaxAddresses(minConfirms, plugin.DefaultMaxConfirms, []btcutil.Address{_addr})
	if err != nil {
		return btcutil.Amount(0), err
	}

	accountAmount := .0
	for _, sp := range unspents {
		if sp.Address == addr {
			accountAmount += sp.Amount
		}
	}

	return btcutil.NewAmount(accountAmount)
}

// ListUnspent ..
func ListUnspent(addr string, minConfirms int) ([]btcjson.ListUnspentResult, error) {
	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	if minConfirms <= 0 {
		minConfirms = plugin.DefaultMinConfirms
	}

	_addr, err := btcutil.DecodeAddress(addr, plugin.BTCNetMap[plugin.CoinNet])
	if err != nil {
		return nil, err
	}

	// TODO gas optimization
	return cli.ListUnspentMinMaxAddresses(minConfirms, plugin.DefaultMaxConfirms, []btcutil.Address{_addr})
}

// SendRawTransaction ..
func SendRawTransaction(rawMsg *wire.MsgTx) (*chainhash.Hash, error) {
	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	return cli.SendRawTransaction(rawMsg, false)
}

// StateSearchMsg ..
func StateSearchMsg(_ctx context.Context, in *chainhash.Hash) (*btcjson.GetTransactionResult, error) {
	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	return cli.GetTransaction(in)
}

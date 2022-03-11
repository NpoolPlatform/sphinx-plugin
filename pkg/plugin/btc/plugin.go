package btc

import (
	"context"
	"errors"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

var ErrWaitMessageOnChainMinConfirms = errors.New("wait message on chain min confirms")

// WalletBalance
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

	coinNet, ok := env.LookupEnv(env.ENVCOINNET)
	if !ok {
		return btcutil.Amount(0), env.ErrEVNCoinNet
	}

	if !plugin.CheckSupportNet(coinNet) {
		return btcutil.Amount(0), env.ErrEVNCoinNetValue
	}

	unspents, err := cli.ListUnspentMin(minConfirms)
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

func ListUnspent(address string, minConf int) ([]btcjson.ListUnspentResult, error) {
	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	if minConf <= 0 {
		minConf = plugin.DefaultMinConfirms
	}

	unspents, err := cli.ListUnspentMin(minConf)
	if err != nil {
		return nil, err
	}

	// TODO gas optimization
	out := make([]btcjson.ListUnspentResult, 0)
	for _, unspent := range unspents {
		if unspent.Address == address {
			out = append(out, unspent)
		}
	}

	return out, nil
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

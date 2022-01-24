package btc

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// WalletBalance
func WalletBalance(accountAddr string, minConfirms int) (btcutil.Amount, error) {
	cli, err := client()
	if err != nil {
		return btcutil.Amount(0), err
	}
	defer cli.Shutdown()

	if minConfirms <= 0 {
		minConfirms = plugin.DefaultMinConfirms
	}

	address, err := btcutil.DecodeAddress(accountAddr, &chaincfg.MainNetParams)
	if err != nil {
		return btcutil.Amount(0), err
	}

	return cli.GetReceivedByAddressMinConf(address, minConfirms)
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

// SendRawTransaction
func SendRawTransaction(rawMsg *wire.MsgTx) (*chainhash.Hash, error) {
	cli, err := client()
	if err != nil {
		return nil, err
	}
	defer cli.Shutdown()

	return cli.SendRawTransaction(rawMsg, false)
}

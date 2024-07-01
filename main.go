package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc/sign"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
)

func main() {
	createAccount()
}

func main1() {
	os.Setenv("ENV_COIN_NET", "main")
	config.SetENV(&config.ENVInfo{
		LocalWalletAddr:  "3.124.146.218:8732|depc@8d470b6e84a844c1de576b68ba25afacc8f3ebd5fb42b60628d1357b5f458ebc",
		PublicWalletAddr: "3.124.146.218:8732|depc@8d470b6e84a844c1de576b68ba25afacc8f3ebd5fb42b60628d1357b5f458ebc",
	})
	// config.SetENV(&config.ENVInfo{
	//  LocalWalletAddr:  "172.16.31.34:18443|test@test",
	//  PublicWalletAddr: "172.16.31.34:18443|test@test",
	// })

	config.SetENV(&config.ENVInfo{
		LocalWalletAddr:  "172.16.31.34:18743|test@test",
		PublicWalletAddr: "172.16.31.34:18743|test@test",
	})

	// config.SetENV(&config.ENVInfo{
	//  LocalWalletAddr:  "172.21.250.11:18443|test@12345679",
	//  PublicWalletAddr: "172.21.250.11:18443|test@12345679",
	// })
	err := btc.Client().WithClient(context.Background(), func(c *rpcclient.Client) (bool, error) {
		fmt.Println(c.Version())
		return false, nil
	})
	fmt.Println(err)
	fmt.Println("sss")
}

func createAccount() {
	netType := &chaincfg.RegressionNetParams
	// secret, err := btcec.NewPrivateKey(btcec.S256())
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// wif, err := btcutil.NewWIF(secret, netType, true)

	wifStr := "cQ4yrDokKGFWfaJujp3HKPoWNqn5QTHYjnAV1JxEt6qRVhhtmzar"
	wif, err := btcutil.DecodeWIF(wifStr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	addressPubKey, err := btcutil.NewAddressPubKey(
		wif.PrivKey.PubKey().SerializeCompressed(),
		netType,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pkscript, err := sign.PayToPubKeyScript(addressPubKey.ScriptAddress())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pksh, err := btcutil.NewAddressScriptHash(pkscript, netType)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	addr := pksh.EncodeAddress()

	fmt.Println(addr)
}

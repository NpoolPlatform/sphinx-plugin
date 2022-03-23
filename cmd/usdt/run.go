package usdt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/eth/usdt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	cli "github.com/urfave/cli/v2"
)

var DeployUSDTCmd = &cli.Command{
	Name:   "usdt-erc20",
	Usage:  "deploy usdt erc20 contract",
	Hidden: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "addr",
			Usage:       "eth server http listen address",
			Hidden:      true,
			DefaultText: "localhost",
		},
		&cli.StringFlag{
			Name:        "port",
			Usage:       "eth server http listen port",
			Hidden:      true,
			DefaultText: "8545",
		},
	},
	Action: func(c *cli.Context) error {
		log.SetFlags(log.Lshortfile)
		// TODO think how to deploy contract
		// create conn
		ctx, cancel := context.WithTimeout(c.Context, time.Minute*5)
		defer cancel()

		// http://localhost:8545
		gethServer := "http://" + net.JoinHostPort(c.String("addr"), c.String("port"))

		log.Printf("create geth client: %s\n", gethServer)

		client, err := ethclient.DialContext(ctx, gethServer)
		if err != nil {
			log.Fatalf("conn eth client err: %v", err)
		}
		defer client.Close()

		// coinbase
		coinbase, err := execGeth(gethServer, "eth.coinbase")
		if err != nil {
			log.Fatalf("get geth coinbase err: %v", err)
		}

		// TODO default wallet(maybe no wallet)
		_defaultWallet, err := execGeth(gethServer, "personal.listWallets[0].accounts[0]")
		if err != nil {
			log.Fatalf("get geth coinbase err: %v", err)
		}

		/*return info not a json
		{
		  address: "0x4f91323c16be425cc8c49b20b83bf19149c4b73d",
		  url: "keystore:///opt/geth/node0/keystore/UTC--2022-03-10T04-19-46.159947188Z--4f91323c16be425cc8c49b20b83bf19149c4b73d"
		}
		*/

		// format cmd get data
		coinbase = bytes.TrimFunc(coinbase, func(r rune) bool {
			if unicode.IsSpace(r) || r == '"' {
				return true
			}
			return false
		})
		_defaultWallet = bytes.Replace(_defaultWallet, []byte("address"), []byte(`"address"`), 1)
		_defaultWallet = bytes.Replace(_defaultWallet, []byte("url"), []byte(`"url"`), 1)

		defaultWallet := &struct {
			Address string `json:"address"`
			URL     string `json:"url"`
		}{}

		if err := json.Unmarshal(_defaultWallet, defaultWallet); err != nil {
			log.Fatalf("get geth wallet err: %v", err)
		}

		if !bytes.Equal(coinbase, []byte(defaultWallet.Address)) {
			log.Fatal("compare geth wallet can only use coinbase wallet")
		}

		walletPath := strings.Split(defaultWallet.URL, "//")
		wallet, err := os.Open(walletPath[1])
		if err != nil {
			log.Fatalf("open wallet err: %v", err)
		}
		defer wallet.Close()

		chainID, err := client.NetworkID(ctx)
		if err != nil {
			log.Fatalf("get eth chainID err: %v", err)
		}

		// only support testnet
		if chainID.Cmp(big.NewInt(1337)) != 0 {
			log.Fatal("only support testnet")
		}

		// get private key
		auth, err := bind.NewTransactorWithChainID(wallet, "", chainID)
		if err != nil {
			log.Fatalf("conn eth client err: %v", err)
		}

		addr, _, _, err := usdt.DeployTetherToken(auth, client, big.NewInt(39822488242032626), "Tether USD", "USDT", big.NewInt(6))
		if err != nil {
			log.Fatalf("decode eth private key err: %v", err)
		}

		color.Green("Tether USDT Contract: %v", addr.Hex())
		return nil
	},
}

func execGeth(addr, cmdArgs string) ([]byte, error) {
	geth, err := exec.LookPath("geth")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s attach %s --exec '%s'", geth, addr, cmdArgs))
	return cmd.CombinedOutput()
}

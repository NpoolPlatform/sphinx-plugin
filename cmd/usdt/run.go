package usdt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/usdt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/fbsobreira/gotron-sdk/pkg/keystore"
	cli "github.com/urfave/cli/v2"
)

var USDTCmd = []*cli.Command{
	{
		Name:    "deploy",
		Aliases: []string{"d"},
		Usage:   "deploy usdt erc20 contract",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Usage: "eth server http listen address",
				// Hidden:      true,
				DefaultText: "localhost",
			},
			&cli.StringFlag{
				Name:  "port",
				Usage: "eth server http listen port",
				// Hidden:      true,
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
	},
	{
		Name:    "balance",
		Aliases: []string{"b"},
		Usage:   "get contract address balance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "net",
				Usage:    "gteh net eg: http://127.0.0.1:8545",
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "id",
				Usage:    "geth contract id",
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "address",
				Usage:    "gteh account address",
				Value:    "",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			log.SetFlags(log.Lshortfile)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*50)
			defer cancel()

			client, err := ethclient.DialContext(ctx, c.String("net"))
			if err != nil {
				log.Fatal(err)
			}

			addr := c.String("address")
			contractID := c.String("id")

			if !common.IsHexAddress(addr) {
				log.Fatal(errors.New("invalid address"))
			}
			if !common.IsHexAddress(contractID) {
				log.Fatal(errors.New("invalid contract id"))
			}

			tetherERC20Token, err := usdt.NewTetherToken(common.HexToAddress(contractID), client)
			if err != nil {
				log.Fatal(err)
			}

			decimal, err := tetherERC20Token.Decimals(&bind.CallOpts{
				Pending: true,
				Context: ctx,
			})
			if err != nil {
				log.Fatal(err)
			}

			color.Red("USDT ERC20 Decimals: %v", decimal.Int64())

			balance, err := tetherERC20Token.BalanceOf(
				&bind.CallOpts{
					Pending: true,
					Context: ctx,
				},
				common.HexToAddress(addr),
			)
			if err != nil {
				log.Fatal(err)
			}
			_balance := big.NewFloat(float64(balance.Int64()))
			_balance.Quo(_balance, big.NewFloat(math.Pow10(int(decimal.Int64()))))
			color.Green("Tether USDT Address: %v Balance: %v", addr, _balance.String())
			return nil
		},
	},
	{
		Name:    "transfer",
		Aliases: []string{"t"},
		Usage:   "transfer contract address balance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "net",
				Usage:    "geth net eg: http://127.0.0.1:8545",
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "id",
				Usage:    "geth contract id",
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "to",
				Usage:    "geth account address",
				Value:    "",
				Required: true,
			},
			&cli.Float64Flag{
				Name:     "value",
				Usage:    "geth account amount",
				Value:    0,
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			log.SetFlags(log.Lshortfile)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*50)
			defer cancel()

			gethAddr := c.String("net")
			client, err := ethclient.DialContext(ctx, gethAddr)
			if err != nil {
				log.Fatal(err)
			}

			// coinbase
			coinbase, err := execGeth(gethAddr, "eth.coinbase")
			if err != nil {
				log.Fatalf("get geth coinbase err: %v", err)
			}

			// TODO default wallet(maybe no wallet)
			_defaultWallet, err := execGeth(gethAddr, "personal.listWallets[0].accounts[0]")
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

			// get private key
			json, err := ioutil.ReadFile(walletPath[1])
			if err != nil {
				log.Fatal(err)
			}

			key, err := keystore.DecryptKey(json, "")
			if err != nil {
				log.Fatal(err)
			}

			id := c.String("id")
			to := c.String("to")
			value := c.Float64("value")

			if !common.IsHexAddress(id) {
				log.Fatal(errors.New("invalid contract id"))
			}

			if !common.IsHexAddress(to) {
				log.Fatal(errors.New("invalid address"))
			}

			gasPrice, err := client.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatal(err)
			}

			chainID, err := client.NetworkID(ctx)
			if err != nil {
				log.Fatal(err)
			}

			// only for test net
			if chainID.Cmp(big.NewInt(1)) == 0 {
				log.Fatal(errors.New("only for test net"))
			}

			nonce, err := client.PendingNonceAt(
				ctx,
				common.HexToAddress(defaultWallet.Address),
			)
			if err != nil {
				log.Fatal(err)
			}

			_abi, err := abi.JSON(strings.NewReader(usdt.TetherTokenABI))
			if err != nil {
				log.Fatal(err)
			}

			amount, ok := big.NewInt(0).SetString(big.NewFloat(value).Mul(big.NewFloat(value), big.NewFloat(math.Pow10(6))).Text('f', 0), 10)
			if !ok {
				log.Fatal(errors.New("invalid amount"))
			}

			toAddr := common.HexToAddress(to)
			input, err := _abi.Pack("transfer",
				toAddr,
				amount,
			)
			if err != nil {
				log.Fatal(err)
			}

			caddr := common.HexToAddress(id)
			baseTx := &types.LegacyTx{
				To:       &caddr,
				Nonce:    nonce,
				GasPrice: gasPrice,
				Gas:      300000,
				Value:    big.NewInt(0),
				Data:     input,
			}

			signedTx, err := types.SignNewTx(key.PrivateKey, types.NewEIP155Signer(chainID), baseTx)
			if err != nil {
				log.Fatal(err)
			}

			if err := client.SendTransaction(ctx, signedTx); err != nil {
				log.Fatal(err)
			}

			state := 0
			for range time.NewTicker(1 * time.Second).C {
				_, isPending, err := client.TransactionByHash(ctx, signedTx.Hash())
				if err != nil {
					log.Fatal(err)
				}

				if isPending {
					continue
				}

				receipt, err := client.TransactionReceipt(ctx, signedTx.Hash())
				if err != nil {
					log.Fatal(err)
				}

				state = int(receipt.Status)
				break
			}

			color.Green("contract id: %v to: %v value: %v transaction id: %v state: %v", id, to, value, signedTx.Hash(), state)
			return nil
		},
	},
}

func execGeth(addr, cmdArgs string) ([]byte, error) {
	geth, err := exec.LookPath("geth")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s --exec '%s' attach %s", geth, cmdArgs, addr))
	return cmd.CombinedOutput()
}

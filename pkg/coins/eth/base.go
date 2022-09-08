package eth

import (
	"context"
	"fmt"
	"strings"
	"time"

	bc_client "github.com/NpoolPlatform/build-chain/pkg/client/v1"
	build_chain "github.com/NpoolPlatform/build-chain/pkg/coins/eth"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

const (
	gasToLow      = `intrinsic gas too low`
	fundsToLow    = `insufficient funds for gas * price + value`
	nonceToLow    = `nonce too low`
	AmountInvalid = `invalid amount`
	TokenToLow    = `token funds to low`
	dialTimeout   = 3 * time.Second
)

var (
	stopErrMsg = []string{gasToLow, fundsToLow, nonceToLow, AmountInvalid, TokenToLow}

	ethTokens = []coins.TokenInfo{
		{Waight: 100, OfficialName: "Ethereum", Decimal: 18, Unit: "ETH", Name: string(coins.Ethereum), TokenType: coins.Ethereum, OfficialContract: string(coins.Ethereum), CoinType: sphinxplugin.CoinType_CoinTypeethereum},
		{Waight: 100, OfficialName: "Tether USD", Decimal: 6, Unit: "USDT", Name: "usdterc20", TokenType: coins.Erc20, OfficialContract: "0xdAC17F958D2ee523a2206206994597C13D831ec7", CoinType: sphinxplugin.CoinType_CoinTypeethereum},
		// TODO: will change it to erc20 tokentype
		{Waight: 100, OfficialName: "Coins USD", Decimal: 6, Unit: "USDC", Name: "usdcerc20", TokenType: coins.USDC, OfficialContract: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", CoinType: sphinxplugin.CoinType_CoinTypeusdcerc20},
	}
)

func init() {
	for i := range ethTokens {
		ethTokens[i].Net = coins.CoinNetMain
		ethTokens[i].Contract = ethTokens[i].OfficialContract
		register.RegisteTokenInfo(&ethTokens[i])
	}

	register.RegisteTokenNetHandler(sphinxplugin.CoinType_CoinTypetethereum, netHandle)
}

func netHandle(tokenInfos []*coins.TokenInfo) error {
	ctx := context.Background()
	bcServer, ok := env.LookupEnv(env.ENVBUIILDCHIANSERVER)
	if !ok {
		return env.ErrENVBuildChainServerNotFound
	}

	bcConn, bcConnErr := bc_client.NewClientConn(ctx, bcServer)
	if bcConnErr != nil {
		return fmt.Errorf("connect server faild, %v", bcConnErr)
	}

	for _, tokenInfo := range tokenInfos {
		if tokenInfo.TokenType == coins.Erc20 {
			go func() {
				_tokenInfo, err := build_chain.CrawlOne(ctx, bcConn, tokenInfo.OfficialContract, false)
				if err != nil {
					return
				}

				tokenInfo.Contract = _tokenInfo.PrivateContract
				tokenInfo.DisableRegiste = false
			}()
			// prevent to be baned
			time.Sleep(build_chain.CrawlInterval)
		} else {
			tokenInfo.DisableRegiste = false
		}
	}

	return nil
}

func TxFailErr(err error) bool {
	if err == nil {
		return false
	}

	for _, v := range stopErrMsg {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}

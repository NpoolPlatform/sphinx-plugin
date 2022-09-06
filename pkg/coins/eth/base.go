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
)

const (
	gasToLow    = `intrinsic gas too low`
	fundsToLow  = `insufficient funds for gas * price + value`
	nonceToLow  = `nonce too low`
	dialTimeout = 3 * time.Second
)

var (
	stopErrMsg = []string{gasToLow, fundsToLow, nonceToLow}

	ethTokens = []coins.TokenInfo{
		{Waight: 100, OfficialName: "Ethereum", Decimal: 18, Unit: "ETH", Name: "ethereum", TokenType: coins.Ethereum, OfficialContract: "ethereum"},
		{Waight: 100, OfficialName: "Tether USD", Decimal: 6, Unit: "USDT", Name: "usdterc20", TokenType: coins.Erc20, OfficialContract: "0xdAC17F958D2ee523a2206206994597C13D831ec7"},
		{Waight: 100, OfficialName: "Coins USD", Decimal: 6, Unit: "USDC", Name: "usdcerc20", TokenType: coins.Erc20, OfficialContract: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
	}
)

func init() {
	for i := range ethTokens {
		ethTokens[i].Net = coins.CoinNetMain
		ethTokens[i].Contract = ethTokens[i].OfficialContract
		ethTokens[i].CoinType = sphinxplugin.CoinType_CoinTypeethereum
		register.RegisteTokenInfo(&ethTokens[i])
	}

	register.RegisteTokenNetHandler(sphinxplugin.CoinType_CoinTypetethereum, netHandle)
}

func netHandle(tokenInfos []*coins.TokenInfo) error {
	ctx := context.Background()
	bcConn, bcConnErr := bc_client.NewClientConn("192.168.49.1:50491")

	for _, tokenInfo := range tokenInfos {
		if tokenInfo.TokenType == coins.Erc20 {
			if bcConnErr != nil {
				return fmt.Errorf("connect server faild, %v", bcConnErr)
			}
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

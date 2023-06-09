package eth

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	bc_client "github.com/NpoolPlatform/build-chain/pkg/client/v1"
	build_chain "github.com/NpoolPlatform/build-chain/pkg/coins/eth"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	v1 "github.com/NpoolPlatform/message/npool/basetypes/v1"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/shopspring/decimal"
)

const (
	GasTooLow           = `intrinsic gas too low`
	FundsTooLow         = `insufficient funds for gas * price + value`
	NonceTooLow         = `nonce too low`
	AmountInvalid       = `invalid amount`
	TokenTooLow         = `token funds too low`
	GetInfoFailed       = `get info failed from the eth node`
	DialTimeout         = 3 * time.Second
	EthExp              = -18
	GasTolerance        = 1.25
	ChainType           = sphinxplugin.ChainType_Ethereum
	ChainNativeUnit     = "ETH"
	ChainAtomicUnit     = "Wei"
	ChainUnitExp        = 18
	ChainNativeCoinName = "ethereum"
	ChainID             = "1"
)

var (
	stopErrMsg = []string{GasTooLow, FundsTooLow, NonceTooLow, AmountInvalid, TokenTooLow}

	ethTokens = []coins.TokenInfo{
		{Waight: 100, OfficialName: "Ethereum", Decimal: 18, Unit: "ETH", Name: ChainNativeCoinName, TokenType: coins.Ethereum, OfficialContract: ChainNativeCoinName, CoinType: sphinxplugin.CoinType_CoinTypeethereum},
		{Waight: 100, OfficialName: "Tether USD", Decimal: 6, Unit: "USDT", Name: "usdterc20", TokenType: coins.Erc20, OfficialContract: "0xdAC17F958D2ee523a2206206994597C13D831ec7", CoinType: sphinxplugin.CoinType_CoinTypeethereum},
		// TODO: will change it to erc20 tokentype
		{Waight: 100, OfficialName: "Coins USD", Decimal: 6, Unit: "USDC", Name: "usdcerc20", TokenType: coins.USDC, OfficialContract: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", CoinType: sphinxplugin.CoinType_CoinTypeusdcerc20},
	}
)

func init() {
	for i := range ethTokens {
		// set chain info
		ethTokens[i].ChainType = ChainType
		ethTokens[i].ChainNativeUnit = ChainNativeUnit
		ethTokens[i].ChainAtomicUnit = ChainAtomicUnit
		ethTokens[i].ChainUnitExp = ChainUnitExp
		ethTokens[i].GasType = v1.GasType_DynamicGas
		ethTokens[i].ChainID = ChainID
		ethTokens[i].ChainNickname = ChainType.String()
		ethTokens[i].ChainNativeCoinName = ChainNativeCoinName

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
		logger.Sugar().Error(bcConnErr)
		return fmt.Errorf("connect server failed, %v", bcConnErr)
	}

	for _, tokenInfo := range tokenInfos {
		if tokenInfo.TokenType == coins.Erc20 {
			go func(token *coins.TokenInfo) {
				_tokenInfo, err := build_chain.CrawlOne(ctx, bcConn, tokenInfo.OfficialContract, false)
				if err != nil {
					logger.Sugar().Error(err)
					return
				}

				token.Contract = _tokenInfo.PrivateContract
				token.DisableRegiste = false
			}(tokenInfo)

			// prevent to be baned
			// time.Sleep(coins.SyncTime[tokenInfo.CoinType])
			time.Sleep(time.Second)
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

func ToEth(value *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(value, EthExp)
}

func ToWei(value float64) (*big.Int, bool) {
	wei := big.NewFloat(0).Mul(big.NewFloat(value), big.NewFloat(math.Pow10(-EthExp)))
	return big.NewInt(0).SetString(wei.Text('f', 0), 10)
}

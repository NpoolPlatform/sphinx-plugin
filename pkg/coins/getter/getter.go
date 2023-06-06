package getter

import (
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"

	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/erc20"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/eth"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/usdc/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/usdc/sign"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol/sign"

	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc/sign"

	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil/sign"

	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/sign"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/trc20/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/trc20/sign"

	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/busd/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/busd/sign"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/sign"
)

func GetTokenInfo(name string) *coins.TokenInfo {
	_tokenInfo, ok := register.NameToTokenInfo[name]
	if !ok {
		return nil
	}
	return _tokenInfo
}

func GetTokenInfos(coinType sphinxplugin.CoinType) map[string]*coins.TokenInfo {
	tokenInfos, ok := register.TokenInfoMap[coinType]
	if !ok {
		return nil
	}
	return tokenInfos
}

func GetTokenHandler(tokenType coins.TokenType, op register.OpType) (register.HandlerDef, error) {
	if _, ok := register.TokenHandlers[tokenType]; !ok {
		return nil, register.ErrTokenHandlerNotExist
	}

	if _, ok := register.TokenHandlers[tokenType][op]; !ok {
		return nil, register.ErrTokenHandlerNotExist
	}
	fn := register.TokenHandlers[tokenType][op]
	return fn, nil
}

func GetTokenNetHandler(coinType sphinxplugin.CoinType) (register.NetHandlerDef, error) {
	if _, ok := register.TokenNetHandlers[coinType]; !ok {
		return nil, register.ErrTokenHandlerNotExist
	}
	fn := register.TokenNetHandlers[coinType]
	return fn, nil
}

func nextStop(err error) bool {
	if err == nil {
		return false
	}

	_, ok := register.AbortErrs[err]
	return ok
}

// Abort ..
func Abort(coinType sphinxplugin.CoinType, err error) bool {
	if err == nil {
		return false
	}

	if nextStop(err) {
		return true
	}

	mf, ok := register.AbortFuncErrs[coinType]
	if ok {
		return mf(err)
	}

	return false
}

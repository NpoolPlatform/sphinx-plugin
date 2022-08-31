package getter

import (
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"

	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/erc20"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/eth"

	// register handle
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol/plugin"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol/sign"
)

func GetTokenInfo(name string) *coins.TokenInfo {
	_tokenInfo, ok := register.NameToTokenInfo[name]
	if !ok {
		return nil
	}
	return _tokenInfo
}

func GetTokenInfos(coinType sphinxplugin.CoinType, coinNet string) map[string]*coins.TokenInfo {
	netTokenInfos, ok := register.TokenInfoMap[coinType]
	if !ok {
		return nil
	}
	tokenInfos, ok := netTokenInfos[coinNet]
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

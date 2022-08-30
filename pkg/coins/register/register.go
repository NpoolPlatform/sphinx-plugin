package register

import (
	"context"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
)

// define handler func
type HandlerDef func(ctx context.Context, payload []byte, token *coins.TokenInfo) ([]byte, error)
type OpType int

const (
	OpGetBalance OpType = 0
	OpPreSign    OpType = 1
	OpBroadcast  OpType = 2
	OpSyncTx     OpType = 3
	OpWalletNew  OpType = 20
	OpSign       OpType = 21
)

var (
	ErrTokenHandlerAlreadyExist = errors.New("token handler is already exist")
	ErrTokenHandlerNotExist     = errors.New("token handler is not exist")

	NameToTokenInfo    = make(map[string]*coins.TokenInfo)
	MainContractToName = make(map[string]string)
	TestContractToName = make(map[string]string)
	// cointype -> coinnet -> tokeninfo
	TokenInfoMap  = make(map[sphinxplugin.CoinType]map[string]map[string]*coins.TokenInfo)
	TokenHandlers = make(map[coins.TokenType]map[OpType]HandlerDef)
)

// registe tokeninfos

func RegisteTokenInfos(tokenInfos []*coins.TokenInfo) {
	if len(tokenInfos) == 0 {
		return
	}

	for _, tokenInfo := range tokenInfos {
		RegisteTokenInfo(tokenInfo)
	}
}

func RegisteTokenInfo(tokenInfo *coins.TokenInfo) {
	_tokenInfo := *tokenInfo
	_tokenInfo.CoinType = coins.ToTestCoinType(_tokenInfo.CoinType)
	_tokenInfo.Net = coins.CoinNetTest
	_tokenInfo.Name = fmt.Sprintf("t%v", tokenInfo.Name)
	registeTokenInfo(tokenInfo)
	registeTokenInfo(&_tokenInfo)
}

//  registe tokeninfo
// one contract to one name,contract and name is both unique
// allow to repeated registe,wahgit to decide whether to update
// please submit mainnet tokeninfo
func registeTokenInfo(tokenInfo *coins.TokenInfo) {
	if tokenInfo == nil {
		return
	}

	ContractToName := TestContractToName
	if tokenInfo.Net == coins.CoinNetMain {
		ContractToName = MainContractToName
	}

	// one officialContract to one name
	// check whether the update
	name, ok := ContractToName[tokenInfo.OfficialContract]
	if ok {
		_tokenInfo := NameToTokenInfo[name]
		if ok && _tokenInfo.Waight >= tokenInfo.Waight {
			return
		}
		delete(TokenInfoMap[_tokenInfo.CoinType], name)
		delete(NameToTokenInfo, _tokenInfo.Name)
		delete(TokenInfoMap[tokenInfo.CoinType][_tokenInfo.Net], _tokenInfo.Name)
	}

	// update
	if _, ok = TokenInfoMap[tokenInfo.CoinType]; !ok {
		TokenInfoMap[tokenInfo.CoinType] = make(map[string]map[string]*coins.TokenInfo)
	}

	if _, ok = TokenInfoMap[tokenInfo.CoinType][tokenInfo.Net]; !ok {
		TokenInfoMap[tokenInfo.CoinType][tokenInfo.Net] = make(map[string]*coins.TokenInfo)
	}

	ContractToName[tokenInfo.OfficialContract] = tokenInfo.Name
	TokenInfoMap[tokenInfo.CoinType][tokenInfo.Net][tokenInfo.Name] = tokenInfo
	NameToTokenInfo[tokenInfo.Name] = tokenInfo
}

// TODO: Other chain will support ,so it should move to public pakege
func RegisteTokenHandler(tokenType coins.TokenType, op OpType, fn HandlerDef) {
	if _, ok := TokenHandlers[tokenType]; !ok {
		TokenHandlers[tokenType] = make(map[OpType]HandlerDef)
	}

	if _, ok := TokenHandlers[tokenType][op]; ok {
		panic(ErrTokenHandlerAlreadyExist)
	}
	TokenHandlers[tokenType][op] = fn
}

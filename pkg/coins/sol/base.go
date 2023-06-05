package sol

import (
	"errors"
	"math/big"
	"strings"

	v1 "github.com/NpoolPlatform/message/npool/basetypes/v1"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	solana "github.com/gagliardetto/solana-go"
)

var (
	// EmptyWalletL ..
	EmptyWalletL = big.Int{}
	// EmptyWalletS ..
	EmptyWalletS = big.Float{}
)

const (
	ChainType           = sphinxplugin.ChainType_Solana
	ChainNativeUnit     = "SOL"
	ChainAtomicUnit     = "lamport"
	ChainUnitExp        = 9
	ChainNativeCoinName = "solana"
	ChainID             = "101"
)

var (
	// ErrSolBlockNotFound ..
	ErrSolBlockNotFound = errors.New("not found confirmed block in solana chain")
	// ErrSolSignatureWrong ..
	ErrSolSignatureWrong = errors.New("solana signature is wrong or failed")
)

var (
	SolTransactionFailed = `sol transaction failed`
	lamportsLow          = `Transfer: insufficient lamports`
	stopErrMsg           = []string{lamportsLow, SolTransactionFailed}
	solanaToken          = &coins.TokenInfo{OfficialName: "Solana", Decimal: 9, Unit: "SOL", Name: ChainNativeCoinName, OfficialContract: ChainNativeCoinName, TokenType: coins.Solana}
)

func init() {
	// set chain info
	solanaToken.ChainType = ChainType
	solanaToken.ChainNativeUnit = ChainNativeUnit
	solanaToken.ChainAtomicUnit = ChainAtomicUnit
	solanaToken.ChainUnitExp = ChainUnitExp
	solanaToken.GasType = v1.GasType_GasUnsupported
	solanaToken.ChainID = ChainID
	solanaToken.ChainNickname = ChainType.String()
	solanaToken.ChainNativeCoinName = ChainNativeCoinName

	solanaToken.Waight = 100
	solanaToken.Net = coins.CoinNetMain
	solanaToken.Contract = solanaToken.OfficialContract
	solanaToken.CoinType = sphinxplugin.CoinType_CoinTypesolana
	register.RegisteTokenInfo(solanaToken)
}

func ToSol(larm uint64) *big.Float {
	// Convert lamports to sol:
	return big.NewFloat(0).
		Quo(
			big.NewFloat(0).SetUint64(larm),
			big.NewFloat(0).SetUint64(solana.LAMPORTS_PER_SOL),
		)
}

func ToLarm(value float64) (uint64, big.Accuracy) {
	return big.NewFloat(0).Mul(
		big.NewFloat(0).SetFloat64(value),
		big.NewFloat(0).SetUint64(solana.LAMPORTS_PER_SOL),
	).Uint64()
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

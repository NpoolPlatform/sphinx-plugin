package sol

import (
	"errors"
	"math/big"
	"strings"

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
)

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

func init() {
	solanaToken.Net = coins.CoinNetMain
	solanaToken.Contract = solanaToken.OfficialContract
	solanaToken.CoinType = sphinxplugin.CoinType_CoinTypesolana
	register.RegisteTokenInfo(solanaToken)
}

var solanaToken = &coins.TokenInfo{Waight: 100, OfficialName: "Solana", Decimal: 9, Unit: "SOL", Name: "solana", OfficialContract: "solana", TokenType: coins.Solana}

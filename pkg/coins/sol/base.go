package sol

import (
	"errors"
	"math/big"

	solana "github.com/gagliardetto/solana-go"
)

var EmptyWalletL = big.Int{}

var EmptyWalletS = big.Float{}

var (
	ErrSolBlockNotFound  = errors.New("not found confirmed block in solana chain")
	ErrSolSignatureWrong = errors.New("solana signature is wrong or failed")
)

func ToSol(larm *big.Int) big.Float {
	// Convert lamports to sol:
	sols := new(big.Float).Quo(new(big.Float).SetInt(larm), new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))

	return *sols
}

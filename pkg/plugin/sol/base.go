package sol

import (
	"errors"
	"math/big"

	solana "github.com/gagliardetto/solana-go"
)

var EmptyWalletL = Larmport(0)
var EmptyWalletS = Sol{Value: *new(big.Float).SetFloat64(0)}

type Larmport uint64
type Sol struct {
	Value big.Float
}

var (
	SolErrBlockNotFound = errors.New("not found confirmed block in solana chain")
	SolSignatureErr     = errors.New("solana signature is wrong or failed")
)

func (larm Larmport) ToSol() Sol {
	lamports := new(big.Float).SetUint64(uint64(larm))
	// Convert lamports to sol:
	sols := new(big.Float).Quo(lamports, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))

	return Sol{*sols}
}

func (sol *Sol) String() string {
	return sol.Value.Text('f', 10)
}

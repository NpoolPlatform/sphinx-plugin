package sol

import (
	"context"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func WalletBalance(ctx context.Context, wallet string) (balance Larmport, err error) {

	if wallet == "" {
		return EmptyWalletL, env.ErrAddressInvalid
	}

	pubKey, err := solana.PublicKeyFromBase58(wallet) // serum token

	if err != nil {
		return EmptyWalletL, err
	}
	api, err := client()

	if err != nil {
		return EmptyWalletL, err
	}
	out, err := api.GetBalance(ctx, pubKey, rpc.CommitmentFinalized)
	if err != nil {
		return EmptyWalletL, err
	}
	return Larmport(out.Value), nil
}

func GetRecentBlock(ctx context.Context) (*rpc.GetRecentBlockhashResult, error) {
	api, err := client()
	if err != nil {
		return nil, err
	}
	return api.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
}

package sol

import (
	"context"
	"math/big"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

func WalletBalance(ctx context.Context, wallet string) (balance big.Int, err error) {
	if wallet == "" {
		return EmptyWalletL, env.ErrAddressInvalid
	}

	pubKey, err := solana.PublicKeyFromBase58(wallet)
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
	return *new(big.Int).SetUint64(out.Value), nil
}

func GetRecentBlock(ctx context.Context) (*rpc.GetRecentBlockhashResult, error) {
	api, err := client()
	if err != nil {
		return nil, err
	}
	return api.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
}

func SendTransaction(ctx context.Context, inMsg *sphinxplugin.UnsignedMessage, inSign *sphinxplugin.Signature) (*solana.Signature, error) {
	from, err := solana.PublicKeyFromBase58(inMsg.From)
	if err != nil {
		return nil, err
	}
	to, err := solana.PublicKeyFromBase58(inMsg.To)
	if err != nil {
		return nil, err
	}
	rhash, err := solana.HashFromBase58(inMsg.GetRecentBhash())
	if err != nil {
		return nil, err
	}
	// Build transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				uint64(inMsg.GetValue()*float64(solana.LAMPORTS_PER_SOL)),
				from,
				to,
			).Build(),
		},
		rhash,
		solana.TransactionPayer(from),
	)
	if err != nil {
		return nil, err
	}
	// add signatures
	tx.Signatures = append(tx.Signatures, solana.SignatureFromBytes(inSign.Data))
	err = tx.VerifySignatures()
	if err != nil {
		return nil, ErrSolSignatureWrong
	}
	api, err := client()
	if err != nil {
		return nil, err
	}
	cid, err := api.SendTransaction(ctx, tx)
	return &cid, err
}

// wait message on chain
func StateSearchMsg(signature solana.Signature) (*rpc.TransactionWithMeta, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.WaitMsgOutTimeout)
	defer cancel()
	api, err := client()
	if err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
			// TODO double-spend
			chainMsg, err := api.GetConfirmedTransaction(ctx, signature)
			if chainMsg != nil {
				return chainMsg, nil
			}
			if err != nil {
				return chainMsg, ErrSolBlockNotFound
			}
		}
	}
}

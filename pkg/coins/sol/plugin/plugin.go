package plugin

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol"
	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// here register plugin func
func init() {
	// main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypesolana,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypesolana,
		sphinxproxy.TransactionState_TransactionStateWait,
		preSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypesolana,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		broadcast,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypesolana,
		sphinxproxy.TransactionState_TransactionStateSync,
		syncTx,
	)

	// test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetsolana,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetsolana,
		sphinxproxy.TransactionState_TransactionStateWait,
		preSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetsolana,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		broadcast,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetsolana,
		sphinxproxy.TransactionState_TransactionStateSync,
		syncTx,
	)

	// register err fsm
	err := coins.RegisterAbortErr(
		env.ErrEVNCoinNet,
		env.ErrEVNCoinNetValue,
		env.ErrAddressInvalid,
		env.ErrSignTypeInvalid,
		env.ErrCIDInvalid,
		sol.ErrSolTransactionFailed,
	)
	if err != nil {
		logger.Sugar().Error(err)
	}
}

func walletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.WalletBalanceRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	v, ok := env.LookupEnv(env.ENVCOINNET)
	if !ok {
		return nil, env.ErrEVNCoinNet
	}
	if !coins.CheckSupportNet(v) {
		return nil, env.ErrEVNCoinNetValue
	}

	if info.Address == "" {
		return nil, env.ErrAddressInvalid
	}

	pubKey, err := solana.PublicKeyFromBase58(info.Address)
	if err != nil {
		return nil, err
	}

	api, err := client(ctx)
	if err != nil {
		return nil, err
	}

	bl, err := api.GetBalance(ctx, pubKey, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	balance := sol.ToSol(bl.Value)
	f, exact := balance.Float64()
	if exact != big.Exact {
		logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
	}

	_out := ct.WalletBalanceResponse{
		Balance:    f,
		BalanceStr: balance.String(),
	}

	return json.Marshal(_out)
}

func preSign(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.BaseInfo{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	api, err := client(ctx)
	if err != nil {
		return nil, err
	}

	recentBlockHash, err := api.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	_out := sol.SignMsgTx{
		BaseInfo:        info,
		RecentBlockHash: recentBlockHash.Value.Blockhash.String(),
	}

	return json.Marshal(_out)
}

func broadcast(ctx context.Context, in []byte) (out []byte, err error) {
	info := sol.BroadcastRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(info.Signature))
	if err != nil {
		return nil, err
	}

	err = tx.VerifySignatures()
	if err != nil {
		return nil, sol.ErrSolSignatureWrong
	}

	api, err := client(ctx)
	if err != nil {
		return nil, err
	}

	cid, err := api.SendTransaction(ctx, tx)
	if err != nil {
		sResp := &ct.SyncResponse{}
		sResp.ExitCode = -1
		out, err := json.Marshal(sResp)
		if err != nil {
			return nil, err
		}
		return out, sol.ErrSolTransactionFailed
	}

	// cid, err := api.SendTransaction(ctx, tx)
	// logger.Sugar().Errorf("ssssss%-v,%-v", cid, err)

	_out := ct.SyncRequest{
		TxID: cid.String(),
	}

	return json.Marshal(_out)
}

// syncTx sync transaction status on chain
func syncTx(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.SyncRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	signature, err := solana.SignatureFromBase58(info.TxID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, sconst.WaitMsgOutTimeout)
	defer cancel()

	api, err := client(ctx)
	if err != nil {
		return nil, err
	}

	// TODO double-spend
	chainMsg, err := api.GetTransaction(
		ctx,
		signature,
		&rpc.GetTransactionOpts{
			Encoding:   solana.EncodingBase58,
			Commitment: rpc.CommitmentFinalized,
		})
	if err != nil {
		return nil, sol.ErrSolBlockNotFound
	}

	if chainMsg != nil && chainMsg.Meta.Err != nil {
		sResp := &ct.SyncResponse{}
		sResp.ExitCode = -1
		out, err := json.Marshal(sResp)
		if err != nil {
			return nil, err
		}
		return out, sol.ErrSolTransactionFailed
	}

	if chainMsg != nil && chainMsg.Meta.Err == nil {
		sResp := &ct.SyncResponse{}
		sResp.ExitCode = 0
		out, err := json.Marshal(sResp)
		if err != nil {
			return nil, err
		}
		return out, nil
	}

	return nil, sol.ErrSolBlockNotFound
}

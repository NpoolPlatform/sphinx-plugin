package types

import (
	"context"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

type IPlugin interface {
	WalletBalance(ctx context.Context, req []byte) ([]byte, error)
	PreSign(ctx context.Context, req []byte) ([]byte, error)
	Broadcast(ctx context.Context, req []byte) ([]byte, error)
	SyncTx(ctx context.Context, req []byte) error
}

type ISign interface {
	NewAccount(ctx context.Context, req []byte) ([]byte, error)
	Sign(ctx context.Context, req []byte) ([]byte, error)
}

type BaseInfo struct {
	ENV   string  `json:"env"`
	From  string  `json:"from"`
	To    string  `json:"to"`
	Value float64 `json:"value"`
}

type BroadcastInfo struct {
	TxID string `json:"tx_id"`
}

type SyncInfo struct {
	ExitCode int64 `json:"exit_code"`
}

// plugin
type WalletBalanceRequest struct {
	Address string `json:"address"`
}

type WalletBalanceResponse struct {
	Balance    float64 `json:"balance"`
	BalanceStr string  `json:"balance_str"`
	// Exact      bool    `json:"_"`
}

type CreateTransactionRequest struct {
	CoinType sphinxplugin.CoinType `json:"coin_type"`
	From     string                `json:"from"`
	To       string                `json:"to"`
	Value    float64               `json:"value"`
}

type CreateTransactionResponse struct {
	CID string `json:"cid"`
}

// sign
type NewAccountRequest struct {
	CoinType string `json:"cointype"`
	ENV      string `json:"env"` // main or test
}

type NewAccountResponse struct {
	Address string `json:"address"`
}

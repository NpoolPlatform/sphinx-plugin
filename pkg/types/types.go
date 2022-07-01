package types

import "context"

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
	From  string  `json:"from"`
	To    string  `json:"to"`
	Value float64 `json:"value"`
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

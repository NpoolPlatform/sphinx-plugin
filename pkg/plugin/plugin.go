package plugin

import "context"

type WalletBalanceInfo struct {
	Balance    float64
	BalanceStr string
	Exact      bool
}

type IPlugin interface {
	NewAccount(ctx context.Context) (string, error)
	WalletBalance(ctx context.Context) (WalletBalanceInfo, error)
	PreSign(ctx context.Context) ([]byte, error)
	Sign(ctx context.Context) ([]byte, error)
	Broadcast(ctx context.Context) (string, error)
	SyncTransaction(ctx context.Context) error
}

package plugin

import "context"

type WalletBalanceInfo struct {
	Balance    float64
	BalanceStr string
	Exact      bool
}

type IPlugin interface {
	NewAccount(ctx context.Context, fun interface{}) (string, error)
	WalletBalance(ctx context.Context, fun interface{}) (WalletBalanceInfo, error)

	// pre
	// sign
	// broadcast
	// sync
}

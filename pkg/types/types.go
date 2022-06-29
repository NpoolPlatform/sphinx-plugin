package types

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
	From  string `json:"from"`
	To    string `json:"To"`
	Value string `json:"Value"`
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

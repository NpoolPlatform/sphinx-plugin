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

// sign
type NewAccountRequest struct {
	ENV string `json:"env"` // main or test
}

type NewAccountResponse struct {
	Address string `json:"address"`
}

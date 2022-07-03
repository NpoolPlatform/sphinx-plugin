package sol

import (
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

type SignMsgTx struct {
	ct.BaseInfo     `json:"base"`
	RecentBlockHash string `json:"recent_block_hash"`
}

type BroadcastRequest struct {
	Signature []byte `json:"signature"`
}

package tron

import (
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
)

type SignMsgTx struct {
	Base        ct.BaseInfo               `json:"base"`
	TxExtension *api.TransactionExtention `json:"tx_extension"`
}

type BroadcastRequest struct {
	TxExtension *api.TransactionExtention `json:"tx_extension"`
}

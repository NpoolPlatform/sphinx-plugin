package task

import (
	"context"
	"os"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/signproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/conn"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/fil"
	"github.com/NpoolPlatform/sphinx-proxy/pkg/check"
)

func Plugin() {
	_conn, err := conn.GetGRPCConn(config.GetString(config.KeySphinxProxyAddr))
	if err != nil {
		logger.Sugar().Errorf("call GetGRPCConn error: %v", err)
		os.Exit(1)
	}

	client := signproxy.NewSignProxyClient(_conn)
	r, err := client.ProxyPlugin(context.Background())
	if err != nil {
		logger.Sugar().Errorf("call ProxyPlugin error: %v", err)
		os.Exit(1)
	}

	for {
		// this will block
		rep, err := r.Recv()
		if err != nil {
			logger.Sugar().Errorf("receiver info error: %v", err)
			continue
		}

		logger.Sugar().Infof(
			"sphinx plugin recv info TransactionType: %v CoinType: %v",
			rep.GetTransactionType(),
			rep.GetCoinType(),
		)

		resp := &signproxy.ProxyPluginResponse{
			TransactionType:     rep.GetTransactionType(),
			CoinType:            rep.GetCoinType(),
			TransactionIDInsite: rep.GetTransactionIDInsite(),
		}

		if err := check.CoinType(rep.GetCoinType()); err != nil {
			logger.Sugar().Errorf("check CoinType: %v invalid", rep.GetCoinType())
			goto send
		}
		if err := plugin(rep, resp); err != nil {
			logger.Sugar().Errorf("plugin deal error: %v", err)
		}
	send:
		{
			if err := r.Send(resp); err != nil {
				logger.Sugar().Errorf("send info error: %v", err)
				continue
			}
		}
	}
}

func plugin(rep *signproxy.ProxyPluginRequest, resp *signproxy.ProxyPluginResponse) error {
	switch rep.GetTransactionType() {
	case signproxy.TransactionType_Balance:
		balance, err := fil.WalletBalance(rep.GetAddress())
		if err != nil {
			return err
		}
		resp.Balance = balance
	case signproxy.TransactionType_PreSign:
		nonce, err := fil.MpoolGetNonce(rep.GetAddress())
		if err != nil {
			return err
		}
		resp.Nonce = nonce
	case signproxy.TransactionType_Broadcast:
		cid, err := fil.MpoolPush(rep.GetMessage(), rep.GetSignature())
		if err != nil {
			return err
		}
		resp.CID = cid
	}
	return nil
}

package task

import (
	"context"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/signproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/fil"
	"github.com/NpoolPlatform/sphinx-proxy/pkg/check"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	delayDuration = time.Second * 10
	delay         = make(chan struct{})
	newConn       = make(chan struct{})
	startDeal     = make(chan struct{})
	err           error
	conn          *grpc.ClientConn
	proxyClient   signproxy.SignProxy_ProxyPluginClient
)

func Plugin() {
	go func() {
		newProxyClinet()
	}()

	for {
		select {
		case <-delay:
			time.Sleep(delayDuration)
			newConn <- struct{}{}
		case <-newConn:
			newProxyClinet()
		case <-startDeal:
			handleTran()
		}
	}
}

func newProxyClinet() {
	conn, err = client.GetGRPCConn(config.GetString(config.KeySphinxProxyAddr))
	if err != nil {
		logger.Sugar().Errorf("call GetGRPCConn error: %v", err)
		delay <- struct{}{}
	}
	pClient := signproxy.NewSignProxyClient(conn)
	proxyClient, err = pClient.ProxyPlugin(context.Background())
	if err != nil {
		logger.Sugar().Errorf("call Transaction error: %v", err)
		delay <- struct{}{}
	}
	startDeal <- struct{}{}
}

func handleTran() {
	for {
		// this will block
		rep, err := proxyClient.Recv()
		if err != nil {
			logger.Sugar().Errorf("receiver info error: %v", err)
			if status.Code(err) == codes.Unavailable {
				proxyClient.CloseSend() // nolint
				delay <- struct{}{}
				break
			}
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
			if err := proxyClient.Send(resp); err != nil {
				logger.Sugar().Errorf("send info error: %v", err)
				if status.Code(err) == codes.Unavailable {
					proxyClient.CloseSend() // nolint
					delay <- struct{}{}
					break
				}
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
		resp.Message = rep.GetMessage()
	case signproxy.TransactionType_Broadcast:
		cid, err := fil.MpoolPush(rep.GetMessage(), rep.GetSignature())
		if err != nil {
			return err
		}
		resp.CID = cid
	}
	return nil
}

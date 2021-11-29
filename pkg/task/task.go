package task

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/signproxy"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/fil"
	"github.com/NpoolPlatform/sphinx-proxy/pkg/check"
	"github.com/NpoolPlatform/sphinx-proxy/pkg/unit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	err                  error
	delayDuration        = time.Second * 2
	registerCoinDuration = time.Second * 5

	// use atomic 1: conn not valid, should renew one
	closeNumFlag int32
	closeConn    = make(chan struct{})
	delay        = make(chan struct{})
	newConn      = make(chan struct{})
	startDeal    = make(chan struct{})
	sendChannel  = make(chan *signproxy.ProxyPluginResponse)
	conn         *grpc.ClientConn
	proxyClient  signproxy.SignProxy_ProxyPluginClient
)

func Plugin() {
	go newProxyClinet()
	for {
		select {
		case <-time.After(registerCoinDuration):
			go func() {
				sendChannel <- &signproxy.ProxyPluginResponse{
					CoinType:        sphinxplugin.CoinType_CoinTypeFIL,
					TransactionType: signproxy.TransactionType_RegisterCoin,
				}
			}()
		case <-delay:
			time.Sleep(delayDuration)
			go func() { newConn <- struct{}{} }()
		case <-newConn:
			go newProxyClinet()
		case <-startDeal:
			go watch()
			go send()
			go recv()
		}
	}
}

func newProxyClinet() {
	logger.Sugar().Info("start new plugin client")
	conn, err = client.GetGRPCConn(config.GetString(config.KeySphinxProxyAddr))
	if err != nil {
		logger.Sugar().Errorf("call GetGRPCConn error: %v", err)
		delay <- struct{}{}
		return
	}
	pClient := signproxy.NewSignProxyClient(conn)
	proxyClient, err = pClient.ProxyPlugin(context.Background())
	if err != nil {
		logger.Sugar().Errorf("call Transaction error: %v", err)
		delay <- struct{}{}
		return
	}
	logger.Sugar().Info("start new plugin client ok")
	startDeal <- struct{}{}
}

func watch() {
	for {
		<-closeConn
		atomic.StoreInt32(&closeNumFlag, 1)
		break
	}
}

func recv() {
	for atomic.LoadInt32(&closeNumFlag) == 0 {
		// this will block
		req, err := proxyClient.Recv()
		if err != nil {
			logger.Sugar().Errorf("receiver info error: %v", err)
			if checkCode(err) {
				proxyClient.CloseSend() // nolint
				delay <- struct{}{}
				break
			}
		}

		logger.Sugar().Infof(
			"sphinx plugin recv info TransactionIDInsite: %v TransactionType: %v CoinType: %v",
			req.GetTransactionIDInsite(),
			req.GetTransactionType(),
			req.GetCoinType(),
		)

		resp := &signproxy.ProxyPluginResponse{
			TransactionType:     req.GetTransactionType(),
			CoinType:            req.GetCoinType(),
			TransactionIDInsite: req.GetTransactionIDInsite(),
			Message:             &sphinxplugin.UnsignedMessage{},
		}

		if err := check.CoinType(req.GetCoinType()); err != nil {
			logger.Sugar().Errorf("check CoinType: %v invalid", req.GetCoinType())
			goto sd
		}

		if err := plugin(req, resp); err != nil {
			logger.Sugar().Errorf("plugin deal error: %v", err)
			goto sd
		}

	sd:
		{
			sendChannel <- resp
		}
	}
}

func send() {
	for atomic.LoadInt32(&closeNumFlag) == 0 {
		resp := <-sendChannel
		if err := proxyClient.Send(resp); err != nil {
			logger.Sugar().Errorf("send info error: %v", err)
			if checkCode(err) {
				proxyClient.CloseSend() // nolint
				delay <- struct{}{}
				break
			}
		}
	}
}

func plugin(req *signproxy.ProxyPluginRequest, resp *signproxy.ProxyPluginResponse) error {
	switch req.GetTransactionType() {
	case signproxy.TransactionType_Balance:
		balance, err := fil.WalletBalance(req.GetAddress())
		if err != nil {
			return err
		}
		f, exist := unit.AttoFIL2FIL(balance.String())
		if exist {
			logger.Sugar().Warnf("wallet balance transfer warning balance: %v", balance.String())
		}
		resp.Balance = f
	case signproxy.TransactionType_PreSign:
		nonce, err := fil.MpoolGetNonce(req.GetAddress())
		if err != nil {
			return err
		}
		resp.Nonce = nonce
		resp.Message = req.GetMessage()
	case signproxy.TransactionType_Broadcast:
		cid, err := fil.MpoolPush(req.GetMessage(), req.GetSignature())
		if err != nil {
			return err
		}
		resp.CID = cid
	}
	return nil
}

func checkCode(err error) bool {
	if err == io.EOF ||
		status.Code(err) == codes.Unavailable ||
		status.Code(err) == codes.Canceled {
		return true
	}
	return false
}

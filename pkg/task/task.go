package task

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/fil"
	"github.com/NpoolPlatform/sphinx-proxy/pkg/check"
	"github.com/shopspring/decimal"
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
	sendChannel  = make(chan *sphinxproxy.ProxyPluginResponse)
	conn         *grpc.ClientConn
	proxyClient  sphinxproxy.SphinxProxy_ProxyPluginClient
)

func Plugin() {
	go newProxyClinet()
	for {
		select {
		case <-time.After(registerCoinDuration):
			logger.Sugar().Info("register new coin")
			go func() {
				sendChannel <- &sphinxproxy.ProxyPluginResponse{
					CoinType:        sphinxplugin.CoinType_CoinTypeFIL,
					TransactionType: sphinxproxy.TransactionType_RegisterCoin,
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
	pClient := sphinxproxy.NewSphinxProxyClient(conn)
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
		logger.Sugar().Info("start watch plugin client")
		<-closeConn
		atomic.StoreInt32(&closeNumFlag, 1)
		logger.Sugar().Info("start watch plugin client exit")
		break
	}
}

func recv() {
	logger.Sugar().Info("plugin client start recv")
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
			"sphinx plugin recv info TransactionID: %v TransactionType: %v CoinType: %v",
			req.GetTransactionID(),
			req.GetTransactionType(),
			req.GetCoinType(),
		)

		resp := &sphinxproxy.ProxyPluginResponse{
			TransactionType: req.GetTransactionType(),
			CoinType:        req.GetCoinType(),
			TransactionID:   req.GetTransactionID(),
			Message:         &sphinxplugin.UnsignedMessage{},
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
	logger.Sugar().Info("plugin client start recv exit")
}

func send() {
	logger.Sugar().Info("plugin client start send")
	for atomic.LoadInt32(&closeNumFlag) == 0 {
		resp := <-sendChannel
		err := proxyClient.Send(resp)
		if err != nil {
			logger.Sugar().Errorf("send info error: %v", err)
			if checkCode(err) {
				proxyClient.CloseSend() // nolint
				delay <- struct{}{}
				break
			}
		}
	}
	logger.Sugar().Info("plugin client start send exit")
}

func plugin(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		balance, err := fil.WalletBalance(req.GetAddress())
		if err != nil {
			return err
		}
		bl, err := decimal.NewFromString(balance.String())
		if err != nil {
			return err
		}
		f, exist := bl.Float64()
		if !exist {
			logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
		}
		resp.Balance = f
		resp.BalanceStr = balance.String()
	case sphinxproxy.TransactionType_PreSign:
		nonce, err := fil.MpoolGetNonce(req.GetAddress())
		if err != nil {
			return err
		}
		resp.Nonce = nonce
		resp.Message = req.GetMessage()
	case sphinxproxy.TransactionType_Broadcast:
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

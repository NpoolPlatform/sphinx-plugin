package task

import (
	"context"
	"io"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/fil"
	"github.com/NpoolPlatform/sphinx-proxy/pkg/check"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	err                  error
	chanBuff             = 100
	delayDuration        = time.Second * 2
	registerCoinDuration = time.Second * 5

	// use atomic 1: conn not valid, should renew one
	closeBadConn = make(chan struct{})
	done         = make(chan struct{})
	delay        = make(chan struct{})
	newConn      = make(chan struct{})
	startDeal    = make(chan struct{})
	sendChannel  = make(chan *sphinxproxy.ProxyPluginResponse, chanBuff)
	conn         *grpc.ClientConn
	proxyClient  sphinxproxy.SphinxProxy_ProxyPluginClient
)

func Plugin() {
	deamon := make(chan struct{})
	go func() {
		for {
			select {
			case <-delay:
				go func() {
					time.Sleep(delayDuration)
					newConn <- struct{}{}
				}()
			case <-newConn:
				go newProxyClient()
			case <-startDeal:
				go watch()
				go send()
				go recv()
				go register()
			}
		}
	}()
	go func() {
		newConn <- struct{}{}
	}()
	<-deamon
}

func newProxyClient() {
	logger.Sugar().Info("start new plugin client")
	conn, err = client.GetGRPCConn(config.GetString(config.KeySphinxProxyAddr))
	if err != nil {
		logger.Sugar().Errorf("call GetGRPCConn error: %v", err)
		closeBadConn <- struct{}{}
		return
	}
	pClient := sphinxproxy.NewSphinxProxyClient(conn)
	proxyClient, err = pClient.ProxyPlugin(context.Background())
	if err != nil {
		logger.Sugar().Errorf("call Transaction error: %v", err)
		closeBadConn <- struct{}{}
		return
	}
	logger.Sugar().Info("start new plugin client ok")
	startDeal <- struct{}{}
}

func watch() {
	logger.Sugar().Info("start watch plugin client")
	<-closeBadConn
	close(done)

	// close invalid conn
	if conn != nil {
		conn.Close()
	}

	// create new conn
	delay <- struct{}{}
	logger.Sugar().Info("start watch plugin client exit")
}

func register() {
	for {
		select {
		case <-done:
			logger.Sugar().Info("register new coin exit")
			return
		case <-time.After(registerCoinDuration):
			logger.Sugar().Info("register new coin")
			sendChannel <- &sphinxproxy.ProxyPluginResponse{
				CoinType:        sphinxplugin.CoinType_CoinTypeFIL,
				TransactionType: sphinxproxy.TransactionType_RegisterCoin,
			}
		}
	}
}

func recv() {
	logger.Sugar().Info("plugin client start recv")
	for {
		select {
		case <-done:
			logger.Sugar().Info("plugin client start recv exit")
			return
		default:
			// this will block
			req, err := proxyClient.Recv()
			if err != nil {
				logger.Sugar().Errorf("receiver info error: %v", err)
				if checkCode(err) {
					proxyClient.CloseSend() // nolint
					closeBadConn <- struct{}{}
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

			logger.Sugar().Infof(
				"sphinx plugin recv info TransactionID: %v TransactionType: %v CoinType: %v done",
				req.GetTransactionID(),
				req.GetTransactionType(),
				req.GetCoinType(),
			)

		sd:
			{
				sendChannel <- resp
			}
		}
	}
}

func send() {
	logger.Sugar().Info("plugin client start send")
	for {
		select {
		case <-done:
			logger.Sugar().Info("plugin client start send exit")
			return
		default:
			// paral deal
			go func(resp *sphinxproxy.ProxyPluginResponse) {
				err := proxyClient.Send(resp)
				if err != nil {
					logger.Sugar().Errorf("send info error: %v", err)
					if checkCode(err) {
						proxyClient.CloseSend() // nolint
						closeBadConn <- struct{}{}
					}
				}
			}(<-sendChannel)
		}
	}
}

func plugin(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()
	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		balance, err := fil.WalletBalance(ctx, req.GetAddress())
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
		nonce, err := fil.MpoolGetNonce(ctx, req.GetAddress())
		if err != nil {
			return err
		}
		resp.Nonce = nonce
		resp.Message = req.GetMessage()
	case sphinxproxy.TransactionType_Broadcast:
		cid, err := fil.MpoolPush(ctx, req.GetMessage(), req.GetSignature())
		if err != nil {
			return err
		}
		resp.CID = cid
	case sphinxproxy.TransactionType_SyncMsgState:
		// TODO 1 find replace cid 2 restry
		msgInfo, err := fil.StateSearchMsg(ctx, req)
		if err != nil {
			if msgInfo != nil {
				// return error code
				resp.ExitCode = int64(msgInfo.Receipt.ExitCode)
			}
			return err
		}
		resp.ExitCode = int64(msgInfo.Receipt.ExitCode)
	}
	return nil
}

func checkCode(err error) bool {
	if err == io.EOF ||
		status.Code(err) == codes.Unavailable ||
		status.Code(err) == codes.Canceled ||
		status.Code(err) == codes.Unimplemented {
		return true
	}
	return false
}

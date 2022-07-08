package task

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/tron"
	"github.com/ethereum/go-ethereum/log"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/busd/plugin" //nolint
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/plugin"      //nolint
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/plugin"      //nolint
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/usdt/plugin" //nolint
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil/plugin"      //nolint
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol/plugin"      //nolint
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron"            //nolint
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/trc20"      //nolint
	trc20 "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/trc20/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/rpc"
	"google.golang.org/grpc"
)

var (
	chanBuff             = 1000
	delayDuration        = time.Second * 2
	registerCoinDuration = time.Second * 5
)

type pluginClient struct {
	closeBadConn chan struct{}
	exitChan     chan struct{}
	sendChannel  chan *sphinxproxy.ProxyPluginResponse

	once        sync.Once
	conn        *grpc.ClientConn
	proxyClient sphinxproxy.SphinxProxy_ProxyPluginClient
}

func Plugin(exitSig chan os.Signal, cleanChan chan struct{}) {
	newClient(exitSig, cleanChan)
}

func newClient(exitSig chan os.Signal, cleanChan chan struct{}) {
	proxyClient := &pluginClient{
		closeBadConn: make(chan struct{}),
		exitChan:     make(chan struct{}),
		sendChannel:  make(chan *sphinxproxy.ProxyPluginResponse, chanBuff),
	}

	conn, pc, err := proxyClient.newProxyClient()
	if err != nil {
		logger.Sugar().Errorf("create new proxy client error: %w", err)
		delayNewClient(exitSig, cleanChan)
		return
	}

	proxyClient.conn, proxyClient.proxyClient = conn, pc

	go proxyClient.watch(exitSig, cleanChan)
	go proxyClient.register()
	go proxyClient.send()
	go proxyClient.recv()
}

func delayNewClient(exitSig chan os.Signal, cleanChan chan struct{}) {
	time.Sleep(delayDuration)
	go newClient(exitSig, cleanChan)
}

func (c *pluginClient) closeProxyClient() {
	c.once.Do(func() {
		logger.Sugar().Info("close plugin conn and client")
		if c != nil {
			close(c.exitChan)
			if c.proxyClient != nil {
				// nolint
				c.proxyClient.CloseSend()
			}
			if c.conn != nil {
				if err := c.conn.Close(); err != nil {
					log.Warn("close conn error: %v", err)
				}
			}
		}
	})
}

func (c *pluginClient) newProxyClient() (*grpc.ClientConn, sphinxproxy.SphinxProxy_ProxyPluginClient, error) {
	logger.Sugar().Info("start new plugin client")
	conn, err := client.GetGRPCConn(config.GetENV().Proxy)
	if err != nil {
		logger.Sugar().Errorf("call GetGRPCConn error: %v", err)
		return nil, nil, err
	}

	pClient := sphinxproxy.NewSphinxProxyClient(conn)
	proxyClient, err := pClient.ProxyPlugin(context.Background())
	if err != nil {
		logger.Sugar().Errorf("call Transaction error: %v", err)
		return nil, nil, err
	}

	logger.Sugar().Info("start new plugin client ok")
	return conn, proxyClient, nil
}

func (c *pluginClient) watch(exitSig chan os.Signal, cleanChan chan struct{}) {
	for {
		select {
		case <-c.closeBadConn:
			logger.Sugar().Info("start watch plugin client")
			<-c.closeBadConn
			c.closeProxyClient()
			logger.Sugar().Info("start watch plugin client exit")
			delayNewClient(exitSig, cleanChan)
		case <-exitSig:
			c.closeProxyClient()
			close(cleanChan)
			return
		}
	}
}

func (c *pluginClient) register() {
	for {
		select {
		case <-c.exitChan:
			logger.Sugar().Info("register new coin exit")
			return
		case <-time.After(registerCoinDuration):
			// TODO coin net
			coinNetwork, coinType, err := env.CoinInfo()
			if err != nil {
				logger.Sugar().Errorf("register new coin error: %v", err)
				continue
			}

			logger.Sugar().Infof("register new coin: %v for %s network", coinType, coinNetwork)
			resp := &sphinxproxy.ProxyPluginResponse{
				CoinType:        coins.CoinStr2CoinType(coinNetwork, coinType),
				TransactionType: sphinxproxy.TransactionType_RegisterCoin,
				ENV:             coinNetwork,
				Unit:            coins.CoinUnit[coins.CoinStr2CoinType(coinNetwork, coinType)],
			}
			c.sendChannel <- resp
		}
	}
}

func (c *pluginClient) recv() {
	logger.Sugar().Info("plugin client start recv")
	for {
		req, err := c.proxyClient.Recv()
		if err != nil {
			logger.Sugar().Errorf("receiver info error: %v", err)
			if rpc.CheckCode(err) {
				c.closeBadConn <- struct{}{}
				break
			}
		}

		go func() {
			coinType := req.GetCoinType()
			transactionType := req.GetTransactionType()
			transactionID := req.GetTransactionID()

			logger.Sugar().Infof(
				"sphinx plugin recv info TransactionID: %v CoinType: %v TransactionType: %v",
				transactionID,
				transactionType,
				coinType,
			)

			now := time.Now()
			defer logger.Sugar().Infof(
				"plugin handle coinType: %v transaction type: %v id: %v use: %v",
				coinType,
				transactionType,
				transactionID,
				time.Since(now).Seconds(),
			)

			handler, err := coins.GetCoinBalancePlugin(coinType, transactionType)
			if err != nil {
				logger.Sugar().Errorf("GetCoinPlugin get handler error: %v", err)
			}
			respPayload, err := handler(context.Background(), req.GetPayload())
			if err != nil {
				logger.Sugar().Errorf("GetCoinPlugin handle deal transaction error: %v", err)
			}

			resp := &sphinxproxy.ProxyPluginResponse{
				TransactionType: req.GetTransactionType(),
				CoinType:        req.GetCoinType(),
				TransactionID:   req.GetTransactionID(),
				Payload:         respPayload,
			}

			c.sendChannel <- resp
		}()
	}
}

func (c *pluginClient) send() {
	logger.Sugar().Info("plugin client start send")
	for {
		select {
		case <-c.exitChan:
			logger.Sugar().Info("plugin client start send exit")
			return
		case resp := <-c.sendChannel:
			err := c.proxyClient.Send(resp)
			if err != nil {
				logger.Sugar().Errorf("send info error: %v", err)
				if rpc.CheckCode(err) {
					c.closeBadConn <- struct{}{}
				}
			}
		}
	}
}

// TODO: remove code under the line
// register coin handle
// nolint
var handleMap = map[sphinxplugin.CoinType]func(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error{
	sphinxplugin.CoinType_CoinTypeusdttrc20:  pluginTRC20,
	sphinxplugin.CoinType_CoinTypetusdttrc20: pluginTRC20,

	sphinxplugin.CoinType_CoinTypetron:  pluginTRX,
	sphinxplugin.CoinType_CoinTypettron: pluginTRX,
}

var _ = func(c *pluginClient, req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) {
	hf, ok := handleMap[req.GetCoinType()]
	if !ok {
		logger.Sugar().Errorf("not register handle for %v", req.GetCoinType())
		resp.RPCExitMessage = fmt.Sprintf("not register handle for %v", req.GetCoinType())
		goto dirct
	}

	if err := hf(req, resp); err != nil {
		logger.Sugar().Errorf("plugin deal error: %v", err)
		resp.RPCExitMessage = err.Error()
		goto dirct
	}

	logger.Sugar().Infof(
		"sphinx plugin recv info TransactionID: %v TransactionType: %v CoinType: %v done",
		req.GetTransactionID(),
		req.GetTransactionType(),
		req.GetCoinType(),
	)

dirct:
	c.sendChannel <- resp
}

// nolint
func pluginTRC20(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		bl, err := trc20.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}

		f := tron.TRC20ToBigFloat(bl)
		resp.Balance, _ = f.Float64()
		resp.BalanceStr = f.Text('f', tron.TRC20ACCURACY)
		// case sphinxplugin.TransactionType_PreSign:
		// 	txExtension, err := trc20.BuildTransaciton(ctx, req)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	txData, err := json.Marshal(txExtension)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	resp.Message = req.GetMessage()
		// 	if resp.GetMessage() == nil {
		// 		resp.Message = &sphinxplugin.UnsignedMessage{}
		// 	}
		// 	resp.Message.TxData = txData
		// case sphinxplugin.TransactionType_Broadcast:
		// 	tx := &api.TransactionExtention{}
		// 	txData := req.GetMessage().GetTxData()
		// 	err := json.Unmarshal(txData, tx)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	err = tron.BroadcastTransaction(ctx, tx.Transaction)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	resp.CID = common.BytesToHexString(tx.GetTxid())
		// case sphinxplugin.TransactionType_SyncMsgState:
		pending, exitcode, err := tron.SyncTxState(ctx, req.GetCID())
		if exitcode == tron.TransactionInfoFAILED {
			resp.ExitCode = exitcode
			return nil
		}

		if err != nil {
			return err
		}
		if !pending {
			return tron.ErrWaitMessageOnChain
		}
	}
	return nil
}

// nolint
func pluginTRX(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		bl, err := tron.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}

		f := tron.TRXToBigFloat(bl)
		resp.Balance, _ = f.Float64()
		resp.BalanceStr = f.Text('f', tron.TRXACCURACY)
		// case sphinxplugin.TransactionType_PreSign:
		// 	txExtension, err := tron.BuildTransaciton(ctx, req)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	txData, err := json.Marshal(txExtension)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	resp.Message = req.GetMessage()
		// 	if resp.GetMessage() == nil {
		// 		resp.Message = &sphinxplugin.UnsignedMessage{}
		// 	}
		// 	resp.Message.TxData = txData
		// case sphinxplugin.TransactionType_Broadcast:
		// 	tx := &api.TransactionExtention{}
		// 	txData := req.GetMessage().GetTxData()
		// 	err := json.Unmarshal(txData, tx)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	err = tron.BroadcastTransaction(ctx, tx.Transaction)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	resp.CID = common.BytesToHexString(tx.GetTxid())
		// case sphinxplugin.TransactionType_SyncMsgState:
		pending, exitcode, err := tron.SyncTxState(ctx, req.GetCID())
		if exitcode == tron.TransactionInfoFAILED {
			resp.ExitCode = exitcode
			return nil
		}

		if err != nil {
			return err
		}
		if !pending {
			return tron.ErrWaitMessageOnChain
		}
	}
	return nil
}

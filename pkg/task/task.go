package task

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"

	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil/plugin"
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

	onec        sync.Once
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
	c.onec.Do(func() {
		logger.Sugar().Info("close plugin conn and client")
		if c != nil {
			close(c.exitChan)
			if c.proxyClient != nil {
				// nolint
				c.proxyClient.CloseSend()
			}
			if c.conn != nil {
				c.conn.Close()
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
			coinType, coinNetwork, err := env.CoinInfo()
			if err != nil {
				logger.Sugar().Errorf("register new coin error: %v", err)
				continue
			}

			logger.Sugar().Infof("register new coin: %v for %s network", coinType, coinNetwork)
			c.sendChannel <- &sphinxproxy.ProxyPluginResponse{
				CoinType:        coins.CoinStr2CoinType(coins.CoinNet, coinType),
				TransactionType: sphinxplugin.TransactionType_RegisterCoin,
				ENV:             coinNetwork,
				Unit:            coins.CoinUnit[coins.CoinStr2CoinType(coins.CoinNet, coinType)],
			}
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

			handler, err := coins.GetCoinPlugin(coinType, transactionType)
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

package task

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"

	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/busd/plugin"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/plugin"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc/plugin"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/plugin"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/usdt/plugin"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil/plugin"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol/plugin"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/plugin"
	// register handle
	_ "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/trc20/plugin"

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
		log.Errorf("create new proxy client error: %w", err)
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
		logger.Sugar().Info("close proxy conn and client")
		if c != nil {
			close(c.exitChan)
			if c.proxyClient != nil {
				if err := c.proxyClient.CloseSend(); err != nil {
					logger.Sugar().Warnf("close proxy conn and client error: %v", err)
				}
			}
			if c.conn != nil {
				if err := c.conn.Close(); err != nil {
					log.Warnf("close conn error: %v", err)
				}
			}
		}
	})
}

func (c *pluginClient) newProxyClient() (*grpc.ClientConn, sphinxproxy.SphinxProxy_ProxyPluginClient, error) {
	logger.Sugar().Info("start new proxy client")
	conn, err := client.GetGRPCConn(config.GetENV().Proxy)
	if err != nil {
		log.Errorf("call GetGRPCConn error: %v", err)
		return nil, nil, err
	}

	pClient := sphinxproxy.NewSphinxProxyClient(conn)
	proxyClient, err := pClient.ProxyPlugin(context.Background())
	if err != nil {
		log.Errorf("call Transaction error: %v", err)
		return nil, nil, err
	}

	logger.Sugar().Info("start new proxy client ok")
	return conn, proxyClient, nil
}

func (c *pluginClient) watch(exitSig chan os.Signal, cleanChan chan struct{}) {
	for {
		select {
		case <-c.closeBadConn:
			logger.Sugar().Info("start watch proxy client")
			<-c.closeBadConn
			c.closeProxyClient()
			logger.Sugar().Info("start watch proxy client exit")
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
			log.Info("register new coin exit")
			return
		case <-time.After(registerCoinDuration):
			// TODO coin net
			coinNetwork, coinType, err := env.CoinInfo()
			if err != nil {
				log.Errorf("register new coin error: %v", err)
				continue
			}

			log.Infof("register new coin: %v for %s network", coinType, coinNetwork)
			resp := &sphinxproxy.ProxyPluginResponse{
				CoinType:           coins.CoinStr2CoinType(coinNetwork, coinType),
				TransactionType:    sphinxproxy.TransactionType_RegisterCoin,
				ENV:                coinNetwork,
				Unit:               coins.CoinUnit[coins.CoinStr2CoinType(coinNetwork, coinType)],
				PluginSerialNumber: env.PluginSerialNumber(),
			}
			c.sendChannel <- resp
		}
	}
}

func (c *pluginClient) recv() {
	log.Info("plugin client start recv")
	for {
		req, err := c.proxyClient.Recv()
		if err != nil {
			log.Errorf("receiver info error: %v", err)
			if rpc.CheckCode(err) {
				c.closeBadConn <- struct{}{}
				break
			}
		}

		go func() {
			coinType := req.GetCoinType()
			transactionType := req.GetTransactionType()
			transactionID := req.GetTransactionID()

			log.Infof(
				"sphinx plugin recv info TransactionID: %v CoinType: %v TransactionType: %v",
				transactionID,
				transactionType,
				coinType,
			)

			now := time.Now()
			defer log.Infof(
				"plugin handle coinType: %v transaction type: %v id: %v use: %v",
				coinType,
				transactionType,
				transactionID,
				time.Since(now).String(),
			)

			var resp *sphinxproxy.ProxyPluginResponse
			handler, err := coins.GetCoinBalancePlugin(coinType, transactionType)
			if err != nil {
				log.Errorf("GetCoinPlugin get handler error: %v", err)
				resp = &sphinxproxy.ProxyPluginResponse{
					TransactionType: req.GetTransactionType(),
					CoinType:        req.GetCoinType(),
					TransactionID:   req.GetTransactionID(),
					RPCExitMessage:  err.Error(),
				}
				goto send
			}
			{
				respPayload, err := handler(context.Background(), req.GetPayload())
				if err != nil {
					log.Errorf("GetCoinPlugin handle deal transaction error: %v", err)
					resp = &sphinxproxy.ProxyPluginResponse{
						TransactionType: req.GetTransactionType(),
						CoinType:        req.GetCoinType(),
						TransactionID:   req.GetTransactionID(),
						RPCExitMessage:  err.Error(),
					}
					goto send
				}

				resp = &sphinxproxy.ProxyPluginResponse{
					TransactionType: req.GetTransactionType(),
					CoinType:        req.GetCoinType(),
					TransactionID:   req.GetTransactionID(),
					Payload:         respPayload,
				}
			}

		send:
			c.sendChannel <- resp
		}()
	}
}

func (c *pluginClient) send() {
	log.Info("plugin client start send")
	for {
		select {
		case <-c.exitChan:
			log.Info("plugin client start send exit")
			return
		case resp := <-c.sendChannel:
			err := c.proxyClient.Send(resp)
			if err != nil {
				log.Errorf("send info error: %v", err)
				if rpc.CheckCode(err) {
					c.closeBadConn <- struct{}{}
				}
			}
		}
	}
}

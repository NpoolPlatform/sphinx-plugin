package task

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/eth"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/eth/usdt"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/fil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/filecoin-project/lotus/build"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	chanBuff             = 1000
	newConn              = make(chan struct{})
	delayDuration        = time.Second * 2
	registerCoinDuration = time.Second * 5
)

type pluginClient struct {
	closeBadConn chan struct{}
	done         chan struct{}
	sendChannel  chan *sphinxproxy.ProxyPluginResponse
	conn         *grpc.ClientConn
	proxyClient  sphinxproxy.SphinxProxy_ProxyPluginClient
}

func Plugin() {
	deamon := make(chan struct{})
	go delayNewClient()
	go func() {
		newClient()
	}()
	<-deamon
}

func delayNewClient() {
	for {
		logger.Sugar().Info("start try to create new plugin client")
		<-newConn
		time.Sleep(delayDuration)
		logger.Sugar().Info("start try to create new plugin client end")
		go newClient()
	}
}

func newClient() {
	proxyClient := &pluginClient{
		closeBadConn: make(chan struct{}),
		done:         make(chan struct{}),
		sendChannel:  make(chan *sphinxproxy.ProxyPluginResponse, chanBuff),
	}

	conn, pc, err := proxyClient.newProxyClient()
	if err != nil {
		newConn <- struct{}{}
		return
	}

	proxyClient.conn, proxyClient.proxyClient = conn, pc

	go proxyClient.watch()
	go proxyClient.register()
	go proxyClient.send()
	go proxyClient.recv()
}

func (c *pluginClient) closeProxyClient() {
	logger.Sugar().Info("close plugin conn and client")
	if c != nil {
		close(c.done)
		if c.conn != nil {
			c.conn.Close()
		}
	}
}

func (c *pluginClient) newProxyClient() (*grpc.ClientConn, sphinxproxy.SphinxProxy_ProxyPluginClient, error) {
	logger.Sugar().Info("start new plugin client")
	conn, err := client.GetGRPCConn(config.GetString(config.KeySphinxProxyAddr))
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

func (c *pluginClient) watch() {
	logger.Sugar().Info("start watch plugin client")
	<-c.closeBadConn
	c.closeProxyClient()
	logger.Sugar().Info("start watch plugin client exit")

	newConn <- struct{}{}
}

func (c *pluginClient) register() {
	for {
		select {
		case <-c.done:
			logger.Sugar().Info("register new coin exit")
			return
		case <-time.After(registerCoinDuration):
			// TODO coin net
			coinTypes, coinNetwork, err := env.CoinInfo()
			if err != nil {
				logger.Sugar().Errorf("register new coin error: %v", err)
				continue
			}

			// 支持多币种
			for _, coinType := range strings.Split(coinTypes, ",") {
				logger.Sugar().Infof("register new coin: %v for %s network", coinType, coinNetwork)
				c.sendChannel <- &sphinxproxy.ProxyPluginResponse{
					CoinType:        plugin.CoinStr2CoinType(plugin.CoinNet, coinType),
					TransactionType: sphinxproxy.TransactionType_RegisterCoin,
					ENV:             coinNetwork,
					Unit:            plugin.CoinUnit[plugin.CoinStr2CoinType(plugin.CoinNet, coinType)],
				}
			}
		}
	}
}

func (c *pluginClient) recv() {
	logger.Sugar().Info("plugin client start recv")
	for {
		select {
		case <-c.done:
			logger.Sugar().Info("plugin client start recv exit")
			return
		default:
			// this will block
			req, err := c.proxyClient.Recv()
			if err != nil {
				logger.Sugar().Errorf("receiver info error: %v", err)
				if checkCode(err) {
					c.proxyClient.CloseSend() // nolint
					c.closeBadConn <- struct{}{}
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

			go func() {
				now := time.Now()
				defer func(req *sphinxproxy.ProxyPluginRequest) {
					logger.Sugar().Infof("plugin handle transaction type: %v id: %v use: %v", req.GetTransactionType(), req.GetTransactionID(), time.Since(now).Seconds())
				}(req)
				handle(c, req, resp)
			}()
		}
	}
}

// register coin handle
var handleMap = map[sphinxplugin.CoinType]func(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error{
	sphinxplugin.CoinType_CoinTypefilecoin:  pluginFIL,
	sphinxplugin.CoinType_CoinTypetfilecoin: pluginFIL,

	sphinxplugin.CoinType_CoinTypebitcoin:  pluginBTC,
	sphinxplugin.CoinType_CoinTypetbitcoin: pluginBTC,

	sphinxplugin.CoinType_CoinTypeethereum:  pluginETH,
	sphinxplugin.CoinType_CoinTypetethereum: pluginETH,

	sphinxplugin.CoinType_CoinTypeusdt:  pluginUSDT,
	sphinxplugin.CoinType_CoinTypetusdt: pluginUSDT,
}

func handle(c *pluginClient, req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) {
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

func (c *pluginClient) send() {
	logger.Sugar().Info("plugin client start send")
	for {
		select {
		case <-c.done:
			logger.Sugar().Info("plugin client start send exit")
			return
		default:
			// paral deal
			for resp := range c.sendChannel {
				go func(resp *sphinxproxy.ProxyPluginResponse) {
					err := c.proxyClient.Send(resp)
					if err != nil {
						logger.Sugar().Errorf("send info error: %v", err)
						if checkCode(err) {
							c.proxyClient.CloseSend() // nolint
							c.closeBadConn <- struct{}{}
						}
					}
				}(resp)
			}
		}
	}
}

func pluginFIL(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		bl, err := fil.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}
		balance := big.NewFloat(float64(bl.Int64()))
		balance.Quo(balance, big.NewFloat(math.Pow10(int(build.FilecoinPrecision))))
		f, exact := balance.Float64()
		if exact != big.Exact {
			logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
		}
		resp.Balance = f
		resp.BalanceStr = balance.String()
	case sphinxproxy.TransactionType_PreSign:
		nonce, err := fil.MpoolGetNonce(ctx, req.GetAddress())
		if err != nil {
			return err
		}
		resp.Message = req.GetMessage()
		if resp.GetMessage() == nil {
			resp.Message = &sphinxplugin.UnsignedMessage{}
		}
		resp.Message.Nonce = nonce
	case sphinxproxy.TransactionType_Broadcast:
		cid, err := fil.MpoolPush(ctx, req.GetMessage(), req.GetSignature())
		if err != nil {
			return err
		}
		resp.CID = cid
	case sphinxproxy.TransactionType_SyncMsgState:
		// TODO 1: find replace cid 2: restry
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

// nolint
func pluginBTC(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		balance, err := btc.WalletBalance(req.GetAddress(), plugin.DefaultMinConfirms)
		if err != nil {
			return err
		}
		resp.Balance = balance.ToBTC()
		resp.BalanceStr = balance.String()
	case sphinxproxy.TransactionType_PreSign:
		// get utxo
		unspents, err := btc.ListUnspent(req.GetAddress(), plugin.DefaultMinConfirms)
		if err != nil {
			return err
		}
		resp.Message = req.GetMessage()
		if resp.GetMessage() == nil {
			resp.Message = &sphinxplugin.UnsignedMessage{}
		}
		for _, unspent := range unspents {
			resp.Message.Unspent = append(resp.Message.Unspent, &sphinxplugin.Unspent{
				TxID:          unspent.TxID,
				Vout:          unspent.Vout,
				Address:       unspent.Address,
				Account:       unspent.Account,
				ScriptPubKey:  unspent.ScriptPubKey,
				RedeemScript:  unspent.RedeemScript,
				Amount:        unspent.Amount,
				Confirmations: unspent.Confirmations,
				Spendable:     unspent.Spendable,
			})
		}
	case sphinxproxy.TransactionType_Broadcast:
		msgTx := req.GetMsgTx()
		txIn := make([]*wire.TxIn, 0)
		txOut := make([]*wire.TxOut, 0)

		for _, _txIn := range msgTx.GetTxIn() {
			cHaxh, err := chainhash.NewHash(_txIn.GetPreviousOutPoint().GetHash())
			if err != nil {
				return err
			}
			txIn = append(txIn, &wire.TxIn{
				PreviousOutPoint: wire.OutPoint{
					Hash:  *cHaxh,
					Index: _txIn.GetPreviousOutPoint().GetIndex(),
				},
				SignatureScript: _txIn.GetSignatureScript(),
				Witness:         _txIn.GetWitness(),
				Sequence:        _txIn.GetSequence(),
			})
		}
		for _, _txOut := range msgTx.GetTxOut() {
			txOut = append(txOut, &wire.TxOut{
				Value:    _txOut.GetValue(),
				PkScript: _txOut.GetPkScript(),
			})
		}

		txHash, err := btc.SendRawTransaction(&wire.MsgTx{
			Version:  msgTx.GetVersion(),
			TxIn:     txIn,
			TxOut:    txOut,
			LockTime: msgTx.GetLockTime(),
		})
		if err != nil {
			return err
		}
		resp.CID = txHash.String()
	case sphinxproxy.TransactionType_SyncMsgState:
		txHash, err := chainhash.NewHashFromStr(req.GetCID())
		if err != nil {
			return err
		}
		tranTx, err := btc.StateSearchMsg(ctx, txHash)
		if err != nil {
			return err
		}
		if tranTx.Confirmations < plugin.DefaultMinConfirms {
			return btc.ErrWaitMessageOnChainMinConfirms
		}
	}
	return nil
}

func pluginETH(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		bl, err := eth.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}

		balance := big.NewFloat(float64(bl.Int64()))
		balance.Quo(balance, big.NewFloat(math.Pow10(18)))
		f, exact := balance.Float64()
		if exact != big.Exact {
			logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
		}

		resp.Balance = f
		resp.BalanceStr = balance.String()
	case sphinxproxy.TransactionType_PreSign:
		preSignInfo, err := eth.PreSign(ctx, req.GetCoinType(), req.GetAddress())
		if err != nil {
			return err
		}
		resp.Message = req.GetMessage()
		if resp.GetMessage() == nil {
			resp.Message = &sphinxplugin.UnsignedMessage{}
		}
		resp.Message.ChainID = preSignInfo.ChainID
		resp.Message.Nonce = preSignInfo.Nonce
		resp.Message.GasPrice = preSignInfo.GasPrice
		resp.Message.GasLimit = preSignInfo.GasLimit
	case sphinxproxy.TransactionType_Broadcast:
		txHash, err := eth.SendRawTransaction(ctx, req.GetSignedRawTxHex())
		if err != nil {
			return err
		}
		resp.CID = txHash
	case sphinxproxy.TransactionType_SyncMsgState:
		pending, err := eth.SyncTxState(ctx, req.GetCID())
		if err != nil {
			return err
		}
		if !pending {
			return eth.ErrWaitMessageOnChain
		}
	}
	return nil
}

func pluginUSDT(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		bls, err := usdt.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}
		balance := big.NewFloat(float64(bls.Balance.Int64()))
		balance.Quo(balance, big.NewFloat(math.Pow10(int(bls.Decimal.Int64()))))
		f, exact := balance.Float64()
		if exact != big.Exact {
			logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
		}
		resp.Balance = f
		resp.BalanceStr = balance.String()
	case sphinxproxy.TransactionType_PreSign:
		preSignInfo, err := eth.PreSign(ctx, req.GetCoinType(), req.GetAddress())
		if err != nil {
			return err
		}
		resp.Message = req.GetMessage()
		if resp.GetMessage() == nil {
			resp.Message = &sphinxplugin.UnsignedMessage{}
		}
		resp.Message.ChainID = preSignInfo.ChainID
		resp.Message.Nonce = preSignInfo.Nonce
		resp.Message.GasPrice = preSignInfo.GasPrice
		resp.Message.GasLimit = preSignInfo.GasLimit
	case sphinxproxy.TransactionType_Broadcast:
		txHash, err := eth.SendRawTransaction(ctx, req.GetSignedRawTxHex())
		if err != nil {
			return err
		}
		resp.CID = txHash
	case sphinxproxy.TransactionType_SyncMsgState:
		pending, err := eth.SyncTxState(ctx, req.GetCID())
		if err != nil {
			return err
		}
		if !pending {
			return eth.ErrWaitMessageOnChain
		}
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

package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"sync"
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
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/sol"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/tron/trc20"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/filecoin-project/lotus/build"
	"github.com/gagliardetto/solana-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	logger.Sugar().Info("start try to create new plugin client")
	time.Sleep(delayDuration)
	logger.Sugar().Info("start try to create new plugin client end")
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
				CoinType:        plugin.CoinStr2CoinType(plugin.CoinNet, coinType),
				TransactionType: sphinxproxy.TransactionType_RegisterCoin,
				ENV:             coinNetwork,
				Unit:            plugin.CoinUnit[plugin.CoinStr2CoinType(plugin.CoinNet, coinType)],
			}
		}
	}
}

func (c *pluginClient) recv() {
	logger.Sugar().Info("plugin client start recv")

	for {
		// this will block
		req, err := c.proxyClient.Recv()
		if err != nil {
			logger.Sugar().Errorf("receiver info error: %v", err)
			if checkCode(err) {
				c.closeBadConn <- struct{}{}
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

		go func() {
			now := time.Now()
			defer func(req *sphinxproxy.ProxyPluginRequest) {
				logger.Sugar().Infof("plugin handle transaction type: %v id: %v use: %v", req.GetTransactionType(), req.GetTransactionID(), time.Since(now).Seconds())
			}(req)
			handle(c, req, resp)
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
				if checkCode(err) {
					c.closeBadConn <- struct{}{}
				}
			}
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

	sphinxplugin.CoinType_CoinTypeusdterc20:  pluginUSDT,
	sphinxplugin.CoinType_CoinTypetusdterc20: pluginUSDT,

	sphinxplugin.CoinType_CoinTypesolana:  pluginSOL,
	sphinxplugin.CoinType_CoinTypetsolana: pluginSOL,

	sphinxplugin.CoinType_CoinTypeusdttrc20:  pluginTRC20,
	sphinxplugin.CoinType_CoinTypetusdttrc20: pluginTRC20,
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

func pluginFIL(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		bl, err := fil.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}
		balance, ok := big.NewFloat(0).SetString(bl.String())
		if !ok {
			return errors.New("convert balance string to float64 error")
		}
		balance.Quo(balance, big.NewFloat(float64((build.FilecoinPrecision))))
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
		// take unit
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

		balance, ok := big.NewFloat(0).SetString(bl.String())
		if !ok {
			return errors.New("convert balance string to float64 error")
		}
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
		bl, err := usdt.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}

		balance, ok := big.NewFloat(0).SetString(bl.Balance.String())
		if !ok {
			return errors.New("convert balance string to float64 error")
		}

		balance.Quo(balance, big.NewFloat(math.Pow10(int(bl.Decimal.Int64()))))
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
		resp.Message.ContractID = preSignInfo.ContractID
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

func pluginSOL(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		bl, err := sol.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}
		balance := sol.ToSol(&bl)

		f, exact := balance.Float64()
		if exact != big.Exact {
			logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
		}
		resp.Balance = f
		resp.BalanceStr = balance.String()
	case sphinxproxy.TransactionType_PreSign:
		rhash, err := sol.GetRecentBlock(ctx)
		if err != nil {
			return err
		}
		resp.Message = req.GetMessage()
		if resp.GetMessage() == nil {
			resp.Message = &sphinxplugin.UnsignedMessage{}
		}
		resp.Message.RecentBhash = rhash.Value.Blockhash.String()
	case sphinxproxy.TransactionType_Broadcast:
		cid, err := sol.SendTransaction(ctx, req.GetMessage(), req.GetSignature())
		if err != nil {
			return err
		}
		resp.CID = cid.String()
	case sphinxproxy.TransactionType_SyncMsgState:
		cid, err := solana.SignatureFromBase58(req.CID)
		if err != nil {
			return err
		}
		_, err = sol.StateSearchMsg(cid)

		if err != nil {
			return err
		}
	}
	return nil
}

func pluginTRC20(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxproxy.TransactionType_Balance:
		bl, err := trc20.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}

		f := trc20.ToFloat(bl)
		resp.Balance, _ = f.Float64()
		resp.BalanceStr = f.Text('f', 10)
	case sphinxproxy.TransactionType_PreSign:
		txExtension, err := trc20.TransactionSend(ctx, req)
		if err != nil {
			return err
		}

		txData, err := json.Marshal(txExtension)
		if err != nil {
			return err
		}

		resp.Message = req.GetMessage()
		if resp.GetMessage() == nil {
			resp.Message = &sphinxplugin.UnsignedMessage{}
		}
		resp.Message.TxData = txData
	case sphinxproxy.TransactionType_Broadcast:
		tx := &api.TransactionExtention{}
		txData := req.GetMessage().GetTxData()
		err := json.Unmarshal(txData, tx)
		if err != nil {
			return err
		}

		err = trc20.BroadcastTransaction(ctx, tx.Transaction)
		if err != nil {
			return err
		}
		resp.CID = common.BytesToHexString(tx.GetTxid())
	case sphinxproxy.TransactionType_SyncMsgState:
		pending, err := trc20.SyncTxState(ctx, req.GetCID())
		if err != nil {
			return err
		}
		if !pending {
			return trc20.ErrWaitMessageOnChain
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

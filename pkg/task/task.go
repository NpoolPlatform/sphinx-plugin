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
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc/busd"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth/usdt"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/trc20"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/rpc"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
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
	time.Sleep(delayDuration)
	go newClient(exitSig, cleanChan)
}

func (c *pluginClient) closeProxyClient() {
	c.onec.Do(func() {
		logger.Sugar().Info("close plugin conn and client")
		if c != nil {
			close(c.exitChan)
			if c.proxyClient != nil {
				if err := c.proxyClient.CloseSend(); err != nil {
					logger.Sugar().Warnf("close plugin conn and client error: %v", err)
				}
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

			respPayload, err := coins.GetCoinPlugin(coinType, transactionType)(context.Background(), req.GetPayload())
			if err != nil {
				logger.Sugar().Errorf("", err)
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

// register coin handle
var handleMap = map[sphinxplugin.CoinType]func(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error{
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

	sphinxplugin.CoinType_CoinTypetron:  pluginTRX,
	sphinxplugin.CoinType_CoinTypettron: pluginTRX,

	sphinxplugin.CoinType_CoinTypebinancecoin:  pluginBSC,
	sphinxplugin.CoinType_CoinTypetbinancecoin: pluginBSC,

	sphinxplugin.CoinType_CoinTypebinanceusd:  pluginBEP20,
	sphinxplugin.CoinType_CoinTypetbinanceusd: pluginBEP20,
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

// nolint
func pluginBTC(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxplugin.TransactionType_Balance:
		balance, err := btc.WalletBalance(req.GetAddress(), coins.DefaultMinConfirms)
		if err != nil {
			return err
		}
		resp.Balance = balance.ToBTC()
		// take unit
		resp.BalanceStr = balance.String()
	case sphinxplugin.TransactionType_PreSign:
		// get utxo
		unspents, err := btc.ListUnspent(req.GetAddress(), coins.DefaultMinConfirms)
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
	case sphinxplugin.TransactionType_Broadcast:
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
	case sphinxplugin.TransactionType_SyncMsgState:
		txHash, err := chainhash.NewHashFromStr(req.GetCID())
		if err != nil {
			return err
		}
		tranTx, err := btc.StateSearchMsg(ctx, txHash)
		if err != nil {
			return err
		}
		if tranTx.Confirmations < coins.DefaultMinConfirms {
			return btc.ErrWaitMessageOnChainMinConfirms
		}
	}
	return nil
}

// nolint
func pluginETH(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxplugin.TransactionType_Balance:
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
	case sphinxplugin.TransactionType_PreSign:
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
	case sphinxplugin.TransactionType_Broadcast:
		txHash, err := eth.SendRawTransaction(ctx, req.GetSignedRawTxHex())
		if err != nil {
			return err
		}
		resp.CID = txHash
	case sphinxplugin.TransactionType_SyncMsgState:
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
	case sphinxplugin.TransactionType_Balance:
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
	case sphinxplugin.TransactionType_PreSign:
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
	case sphinxplugin.TransactionType_Broadcast:
		txHash, err := eth.SendRawTransaction(ctx, req.GetSignedRawTxHex())
		if err != nil {
			return err
		}
		resp.CID = txHash
	case sphinxplugin.TransactionType_SyncMsgState:
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
	case sphinxplugin.TransactionType_Balance:
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
	case sphinxplugin.TransactionType_PreSign:
		rhash, err := sol.GetRecentBlock(ctx)
		if err != nil {
			return err
		}
		resp.Message = req.GetMessage()
		if resp.GetMessage() == nil {
			resp.Message = &sphinxplugin.UnsignedMessage{}
		}
		resp.Message.RecentBhash = rhash.Value.Blockhash.String()
	case sphinxplugin.TransactionType_Broadcast:
		cid, err := sol.SendTransaction(ctx, req.GetMessage(), req.GetSignature())
		if err != nil {
			return err
		}
		resp.CID = cid.String()
	case sphinxplugin.TransactionType_SyncMsgState:
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

// nolint
func pluginTRC20(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxplugin.TransactionType_Balance:
		bl, err := trc20.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}

		f := tron.TRC20ToBigFloat(bl)
		resp.Balance, _ = f.Float64()
		resp.BalanceStr = f.Text('f', tron.TRC20ACCURACY)
	case sphinxplugin.TransactionType_PreSign:
		txExtension, err := trc20.BuildTransaciton(ctx, req)
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
	case sphinxplugin.TransactionType_Broadcast:
		tx := &api.TransactionExtention{}
		txData := req.GetMessage().GetTxData()
		err := json.Unmarshal(txData, tx)
		if err != nil {
			return err
		}

		err = tron.BroadcastTransaction(ctx, tx.Transaction)
		if err != nil {
			return err
		}
		resp.CID = common.BytesToHexString(tx.GetTxid())
	case sphinxplugin.TransactionType_SyncMsgState:
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
	case sphinxplugin.TransactionType_Balance:
		bl, err := tron.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			return err
		}

		f := tron.TRXToBigFloat(bl)
		resp.Balance, _ = f.Float64()
		resp.BalanceStr = f.Text('f', tron.TRXACCURACY)
	case sphinxplugin.TransactionType_PreSign:
		txExtension, err := tron.BuildTransaciton(ctx, req)
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
	case sphinxplugin.TransactionType_Broadcast:
		tx := &api.TransactionExtention{}
		txData := req.GetMessage().GetTxData()
		err := json.Unmarshal(txData, tx)
		if err != nil {
			return err
		}

		err = tron.BroadcastTransaction(ctx, tx.Transaction)
		if err != nil {
			return err
		}
		resp.CID = common.BytesToHexString(tx.GetTxid())
	case sphinxplugin.TransactionType_SyncMsgState:
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
func pluginBSC(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxplugin.TransactionType_Balance:
		bl, err := bsc.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			logger.Sugar().Errorf("get balance fail: %v", err)
			return err
		}

		balance, ok := big.NewFloat(0).SetString(bl.String())
		if !ok {
			return errors.New("convert balance string to float64 error")
		}
		balance.Quo(balance, big.NewFloat(math.Pow10(bsc.BNBACCURACY)))
		f, exact := balance.Float64()
		if exact != big.Exact {
			logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
		}

		resp.Balance = f
		resp.BalanceStr = balance.String()
	case sphinxplugin.TransactionType_PreSign:
		preSignInfo, err := bsc.PreSign(ctx, req.GetCoinType(), req.GetAddress())
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
	case sphinxplugin.TransactionType_Broadcast:
		txHash, err := bsc.SendRawTransaction(ctx, req.GetSignedRawTxHex())
		if err != nil {
			return err
		}
		resp.CID = txHash
	case sphinxplugin.TransactionType_SyncMsgState:
		pending, err := bsc.SyncTxState(ctx, req.GetCID())
		if err != nil {
			return err
		}
		if !pending {
			return bsc.ErrWaitMessageOnChain
		}
	}
	return nil
}

// nolint
func pluginBEP20(req *sphinxproxy.ProxyPluginRequest, resp *sphinxproxy.ProxyPluginResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.GrpcTimeout)
	defer cancel()

	switch req.GetTransactionType() {
	case sphinxplugin.TransactionType_Balance:
		bl, err := busd.WalletBalance(ctx, req.GetAddress())
		if err != nil {
			logger.Sugar().Errorf("get balance fail:%v", err)
			return err
		}

		balance, ok := big.NewFloat(0).SetString(bl.String())
		if !ok {
			return errors.New("convert balance string to float64 error")
		}
		balance.Quo(balance, big.NewFloat(math.Pow10(bsc.BEP20ACCURACY)))
		f, exact := balance.Float64()
		if exact != big.Exact {
			logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
		}

		resp.Balance = f
		resp.BalanceStr = balance.String()
	case sphinxplugin.TransactionType_PreSign:
		preSignInfo, err := bsc.PreSign(ctx, req.GetCoinType(), req.GetAddress())
		if err != nil {
			return err
		}
		resp.Message = req.GetMessage()
		if resp.GetMessage() == nil {
			resp.Message = &sphinxplugin.UnsignedMessage{}
		}
		resp.Message.ChainID = preSignInfo.ChainID
		resp.Message.Nonce = preSignInfo.Nonce
		resp.Message.ContractID = preSignInfo.ContractID
		resp.Message.GasPrice = preSignInfo.GasPrice
		resp.Message.GasLimit = preSignInfo.GasLimit
	case sphinxplugin.TransactionType_Broadcast:
		txHash, err := bsc.SendRawTransaction(ctx, req.GetSignedRawTxHex())
		if err != nil {
			return err
		}
		resp.CID = txHash
	case sphinxplugin.TransactionType_SyncMsgState:
		pending, err := bsc.SyncTxState(ctx, req.GetCID())
		if err != nil {
			return err
		}
		if !pending {
			return bsc.ErrWaitMessageOnChain
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

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	pconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

func init() {
	// TODO: support from env or config dynamic set [3,6)
	_interval, ok := env.LookupEnv(env.ENVSYNCINTERVAL)
	if !ok || _interval == "" {
		_coinNet, _coinType, err := env.CoinInfo()
		if err != nil {
			panic(fmt.Sprintf("task::synctx failed to read %v, %v", env.ENVCOINTYPE, err))
		}
		coinType := coins.CoinStr2CoinType(_coinNet, _coinType)
		_interval = strconv.FormatInt(int64(coins.SyncTime[coinType].Seconds()), 10)
	}
	interval, err := strconv.ParseInt(_interval, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("task::synctx failed to read %v, %v", env.ENVSYNCINTERVAL, err))
	}
	if err := register(
		"task::synctx",
		time.Duration(interval*int64(time.Second)),
		syncTx,
	); err != nil {
		fatalf("task::synctx", "task already register")
	}
}

func syncTx(name string, interval time.Duration) {
	for range time.NewTicker(interval).C {
		func() {
			conn, err := client.GetGRPCConn(config.GetENV().Proxy)
			if err != nil {
				errorf(name, "call GetGRPCConn error: %v", err)
				return
			}

			coinNetwork, coinType, err := env.CoinInfo()
			if err != nil {
				errorf(name, "get coin info from env error: %v", err)
				return
			}

			_coinType := coins.CoinStr2CoinType(coinNetwork, coinType)

			tState := sphinxproxy.TransactionState_TransactionStateSync
			handler, err := coins.GetCoinPlugin(
				_coinType,
				tState,
			)
			if err != nil {
				errorf(name, "GetCoinPlugin get handler error: %v", err)
				return
			}

			pClient := sphinxproxy.NewSphinxProxyClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), getTransactionsTimeout)
			ctx = pconst.SetPluginInfo(ctx)
			defer cancel()

			transInfos, err := pClient.GetTransactions(ctx, &sphinxproxy.GetTransactionsRequest{
				ENV:              coinNetwork,
				CoinType:         _coinType,
				TransactionState: tState,
			})
			if err != nil {
				errorf(name, "call Transaction error: %v", err)
				return
			}

			for _, transInfo := range transInfos.GetInfos() {
				func(transInfo *sphinxproxy.TransactionInfo) {
					ctx, cancel := context.WithTimeout(ctx, updateTransactionsTimeout)
					defer cancel()

					now := time.Now()
					defer func() {
						infof(
							name,
							"plugin handle coinType: %v transaction type: %v id: %v use: %v",
							transInfo.GetName(),
							transInfo.GetTransactionState(),
							transInfo.GetTransactionID(),
							time.Since(now).String(),
						)
					}()

					var (
						syncInfo = types.SyncResponse{}
						state    = sphinxproxy.TransactionState_TransactionStateDone
					)

					respPayload, err := handler(ctx, transInfo.GetPayload())
					if err == nil {
						goto done
					}
					if coins.Abort(_coinType, err) {
						errorf(name,
							"sync transaction: %v error: %v stop",
							transInfo.GetTransactionID(),
							err,
						)
						state = sphinxproxy.TransactionState_TransactionStateFail
						goto done
					}

					errorf(name,
						"sync transaction: %v error: %v retry",
						transInfo.GetTransactionID(),
						err,
					)
					return

					// TODO: delete this dirty code
				done:
					{
						if respPayload != nil {
							if err := json.Unmarshal(respPayload, &syncInfo); err != nil {
								errorf(name, "unmarshal sync info error: %v", err)
								return
							}
						}
					}

					if _, err := pClient.UpdateTransaction(ctx, &sphinxproxy.UpdateTransactionRequest{
						TransactionID:        transInfo.GetTransactionID(),
						TransactionState:     tState,
						NextTransactionState: state,
						ExitCode:             syncInfo.ExitCode,
						Payload:              respPayload,
					}); err != nil {
						errorf(name, "UpdateTransaction transaction: %v error: %v", transInfo.GetTransactionID(), err)
						return
					}

					infof(name, "UpdateTransaction transaction: %v done", transInfo.GetTransactionID())
				}(transInfo)
			}
		}()
	}
}

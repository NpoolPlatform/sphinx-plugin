package task

import (
	"context"
	"encoding/json"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/utils"
)

func init() {
	// TODO: support from env or config dynamic set [3,6)
	if err := register(
		"task::synctx",
		config.GetInt(env.ENVSYNCINTERVAL),
		syncTx,
	); err != nil {
		fatalf("task::synctx", "task already register")
	}
}

func syncTx(name string, interval int) {
	for range time.NewTicker(time.Second * time.Duration(interval)).C {
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

			_coinType, err := utils.ToCoinType(coinType)
			if err != nil {
				errorf(name, "transafer coin name error: %v", err)
				return
			}

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
					defer infof(
						name,
						"plugin handle coinType: %v transaction type: %v id: %v use: %v",
						transInfo.GetName(),
						transInfo.GetTransactionID(),
						time.Since(now).Seconds(),
					)

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
							"sync transaction: %v error: %v retry",
							transInfo.GetTransactionID(),
							err,
						)
						state = sphinxproxy.TransactionState_TransactionStateFail
						goto done
					}

					errorf(name,
						"sync transaction: %v error: %v stop",
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

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
	pconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

func init() {
	// TODO: support from env or config dynamic set
	if err := register("task::broadcast", 3*time.Second, broadcast); err != nil {
		fatalf("task::broadcast", "task already register")
	}
}

func broadcast(name string, interval time.Duration) {
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

			tState := sphinxproxy.TransactionState_TransactionStateBroadcast
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
			ctx = pconst.SetPluginSN(ctx, env.PluginSerialNumber())
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
							transInfo.TransactionState,
							transInfo.GetTransactionID(),
							time.Since(now).String(),
						)
					}()

					var (
						broadcastInfo = types.BroadcastInfo{}
						state         = sphinxproxy.TransactionState_TransactionStateSync
					)

					respPayload, err := handler(ctx, transInfo.GetPayload())
					if err == nil {
						goto done
					}
					if coins.Abort(_coinType, err) {
						warnf(name,
							"broadcase transaction: %v error: %v stop",
							transInfo.GetTransactionID(),
							err,
						)
						state = sphinxproxy.TransactionState_TransactionStateFail

						goto done
					}

					errorf(name,
						"broadcase transaction: %v error: %v retry",
						transInfo.GetTransactionID(),
						err,
					)
					return

					// TODO: delete this dirty code
				done:
					{
						if respPayload != nil {
							if err := json.Unmarshal(respPayload, &broadcastInfo); err != nil {
								errorf(name, "unmarshal broadcast info error: %v", err)
							}
						}
					}

					if _, err := pClient.UpdateTransaction(ctx, &sphinxproxy.UpdateTransactionRequest{
						TransactionID:        transInfo.GetTransactionID(),
						TransactionState:     tState,
						NextTransactionState: state,
						CID:                  broadcastInfo.TxID,
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

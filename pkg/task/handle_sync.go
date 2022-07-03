package task

import (
	"context"
	"encoding/json"
	"math/rand"
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
	if err := register("task::synctx", rand.Intn(3)+3, syncTx); err != nil {
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

					respPayload, err := handler(ctx, transInfo.GetPayload())
					if err != nil {
						errorf(name, "GetCoinPlugin handle deal transaction error: %v", err)
						return
					}

					_ = Abort(err)

					// if some error {
					// 	continue retry
					// }

					// TODO: delete this dirty code
					syncInfo := types.SyncResponse{}
					if err := json.Unmarshal(respPayload, &syncInfo); err != nil {
						errorf(name, "unmarshal sync info error: %v", err)
						return
					}

					if _, err := pClient.UpdateTransaction(ctx, &sphinxproxy.UpdateTransactionRequest{
						TransactionID:        transInfo.GetTransactionID(),
						TransactionState:     tState,
						NextTransactionState: sphinxproxy.TransactionState_TransactionStateDone,
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

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
)

func init() {
	// TODO: support from env or config dynamic set
	if err := register("task::nonce", rand.Intn(3)+3, nonce); err != nil {
		fatalf("task::nonce", "task already register")
	}
}

func nonce(name string, interval int) {
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

			_coinType := coins.CoinStr2CoinType(coinNetwork, coinType)

			tState := sphinxproxy.TransactionState_TransactionStateWait
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

					preSignPayload, err := json.Marshal(types.BaseInfo{
						ENV:      coinNetwork,
						CoinType: _coinType,
						From:     transInfo.GetFrom(),
						To:       transInfo.GetTo(),
						Value:    transInfo.GetAmount(),
					})
					if err != nil {
						errorf(name, "marshal presign info error: %v", err)
						return
					}

					state := sphinxproxy.TransactionState_TransactionStateSign

					respPayload, err := handler(ctx, preSignPayload)
					if err == nil {
						goto done
					}

					if coins.Abort(_coinType, err) {
						errorf(name,
							"pre sign transaction: %v error: %v stop",
							transInfo.GetTransactionID(),
							err,
						)
						state = sphinxproxy.TransactionState_TransactionStateFail
						goto done
					}

					errorf(name,
						"pre sign transaction: %v error: %v retry",
						transInfo.GetTransactionID(),
						err,
					)
					return

				done:
					if _, err := pClient.UpdateTransaction(ctx, &sphinxproxy.UpdateTransactionRequest{
						TransactionID:        transInfo.GetTransactionID(),
						TransactionState:     tState,
						NextTransactionState: state,
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

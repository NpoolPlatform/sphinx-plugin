package task

import (
	"context"
	"time"

	grpc2 "github.com/NpoolPlatform/go-service-framework/pkg/grpc"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/signproxy"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	pconst "github.com/NpoolPlatform/sphinx-proxy/pkg/message/const"
)

func RegisterCoin() {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for range tick.C {
		if err := func() error {
			conn, err := grpc2.GetGRPCConn(pconst.ServiceName, grpc2.GRPCTAG)
			if err != nil {
				logger.Sugar().Errorf("call GetGRPCConn error: %v", err)
				return err
			}

			proxyClient := signproxy.NewSignProxyClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err = proxyClient.RegisterCoin(ctx, &signproxy.RegisterCoinRequest{
				CoinType: sphinxplugin.CoinType_CoinTypeFIL,
			})
			if err != nil {
				logger.Sugar().Errorf(
					"call RegisterCoin CoinType: %v error: %v",
					sphinxplugin.CoinType_CoinTypeFIL,
					err,
				)
				return err
			}
			return nil
		}(); err == nil {
			break
		}
	}
}

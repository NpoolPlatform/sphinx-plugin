package task

import (
	"context"
	"os"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/signproxy"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/conn"
)

func RegisterCoin() {
	_conn, err := conn.GetGRPCConn(config.GetString(config.KeySphinxProxyAddr))
	if err != nil {
		logger.Sugar().Errorf("call GetGRPCConn error: %v", err)
		os.Exit(1)
	}

	client := signproxy.NewSignProxyClient(_conn)
	r, err := client.ProxyPlugin(context.Background())
	if err != nil {
		logger.Sugar().Errorf("call ProxyPlugin error: %v", err)
		os.Exit(1)
	}

	for {
		err := r.Send(&signproxy.ProxyPluginResponse{
			TransactionType: signproxy.TransactionType_RegisterCoin,
			CoinType:        sphinxplugin.CoinType_CoinTypeFIL,
		})
		if err != nil {
			logger.Sugar().Errorf("receiver info error: %v", err)
			continue
		}
		r.CloseSend() // nolint
		break
	}
}

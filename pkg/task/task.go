package task

import (
	"context"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/signproxy"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	restryDelay = make(chan struct{})
	next        = make(chan struct{})
	done        = make(chan struct{})
)

func RegisterCoin() {
	go watchRegister()
	registerCoin()
}

func watchRegister() {
lo:
	for {
		select {
		case <-restryDelay:
			time.Sleep(delayDuration)
			next <- struct{}{}
		case <-next:
			registerCoin()
		case <-done:
			break lo
		}
	}
}

func registerCoin() {
	_conn, err := client.GetGRPCConn(config.GetString(config.KeySphinxProxyAddr))
	if err != nil {
		logger.Sugar().Errorf("call GetGRPCConn error: %v", err)
		restryDelay <- struct{}{}
		return
	}

	sProxyClient := signproxy.NewSignProxyClient(_conn)
	r, err := sProxyClient.ProxyPlugin(context.Background())
	if err != nil {
		logger.Sugar().Errorf("call ProxyPlugin error: %v", err)
		restryDelay <- struct{}{}
		return
	}

	err = r.Send(&signproxy.ProxyPluginResponse{
		TransactionType: signproxy.TransactionType_RegisterCoin,
		CoinType:        sphinxplugin.CoinType_CoinTypeFIL,
	})
	if err != nil {
		logger.Sugar().Errorf("receiver info error: %v", err)
		if status.Code(err) == codes.Unavailable {
			r.CloseSend() // nolint
			restryDelay <- struct{}{}
			return
		}
	}
	r.CloseSend() // nolint
}

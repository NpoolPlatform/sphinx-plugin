package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/api"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/task"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
)

var (
	grpcServer *grpc.Server
	runCmd     = &cli.Command{
		Name:    "run",
		Aliases: []string{"s"},
		Usage:   "Run Sphinx Plugin daemon",
		After: func(c *cli.Context) error {
			grpcServer.GracefulStop()
			return logger.Sync()
		},
		Action: func(c *cli.Context) error {
			go task.RegisterCoin()
			return rpcRegister()
		},
	}
)

func rpcRegister() error {
	gport := viper.GetString("grpc_port")
	prometheusPort := viper.GetString("prometheus_port")

	l, err := net.Listen("tcp", fmt.Sprintf(":%v", gport))
	if err != nil {
		return xerrors.Errorf("fail to listen tcp at %v: %v", gport, err)
	}

	grpcServer = grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor,
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
		)),
	)

	api.Register(grpcServer)

	// prometheus metrics endpoints
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(grpcServer)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%v", prometheusPort), nil) //nolint
	}()

	return grpcServer.Serve(l)
}

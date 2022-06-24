package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/task"
	cli "github.com/urfave/cli/v2"
)

var runCmd = &cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Usage:   "Run Sphinx Plugin daemon",
	Before: func(c *cli.Context) error {
		if !coins.CheckSupportNet(coins.CoinNet) {
			// TODO should exit ??
			os.Exit(1)
		}
		return nil
	},
	After: func(c *cli.Context) error {
		return logger.Sync()
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "coin_net",
			Hidden:      true,
			Destination: &coins.CoinNet,
			EnvVars:     []string{env.ENVCOINNET},
		},
	},
	Action: func(c *cli.Context) error {
		sigs := make(chan os.Signal, 1)
		cleanChan := make(chan struct{})
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		task.Plugin(sigs, cleanChan)
		<-cleanChan
		logger.Sugar().Info("graceful shutdown plugin service")
		return nil
	},
}

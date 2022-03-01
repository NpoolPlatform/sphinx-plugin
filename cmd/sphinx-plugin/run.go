package main

import (
	"os"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/btc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/task"
	cli "github.com/urfave/cli/v2"
)

var runCmd = &cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Usage:   "Run Sphinx Plugin daemon",
	Before: func(c *cli.Context) error {
		if !btc.CoinNetMapCheck(plugin.CoinNet) {
			// TODO should exit!!
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
			Destination: &plugin.CoinNet,
			Aliases:     []string{"coin_net"},
			EnvVars:     []string{env.ENVCOINNET},
		},
	},
	Action: func(c *cli.Context) error {
		task.Plugin()
		return nil
	},
}

package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/task"
	cli "github.com/urfave/cli/v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var runCmd = &cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Usage:   "Run Sphinx Plugin daemon",
	After: func(c *cli.Context) error {
		plugin.Close()
		return logger.Sync()
	},
	Action: func(c *cli.Context) error {
		log.Infof(
			"run plugin wanIP: %v, Position %v",
			config.GetENV().WanIP,
			config.GetENV().Position,
		)

		task.Run()
		sigs := make(chan os.Signal, 1)
		cleanChan := make(chan struct{})
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		task.Plugin(sigs, cleanChan)
		<-cleanChan
		log.Info("graceful shutdown plugin service")
		return nil
	},
}

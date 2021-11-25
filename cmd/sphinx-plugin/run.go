package main

import (
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/task"
	cli "github.com/urfave/cli/v2"
)

var runCmd = &cli.Command{
	Name:    "run",
	Aliases: []string{"s"},
	Usage:   "Run Sphinx Plugin daemon",
	After: func(c *cli.Context) error {
		return logger.Sync()
	},
	Action: func(c *cli.Context) error {
		task.Plugin()
		return nil
	},
}

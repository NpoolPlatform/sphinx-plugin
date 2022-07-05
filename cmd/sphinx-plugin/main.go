package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/version"
	"github.com/NpoolPlatform/sphinx-plugin/cmd/usdt"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	banner "github.com/common-nighthawk/go-figure"
	cli "github.com/urfave/cli/v2"
)

const (
	serviceName = "Sphinx Plugin"
	usageText   = "Sphinx Plugin Service"
)

var (
	proxyAddress string
	syncInterval int64
	contract     string
	logDir       string
	logLevel     string
)

func main() {
	commands := cli.Commands{
		runCmd,
		usdt.DeployUSDTCmd,
	}

	description := fmt.Sprintf("%v service cli\nFor help on any individual command run <%v COMMAND -h>\n",
		serviceName, serviceName)

	banner.NewColorFigure(serviceName, "", "green", true).Print()
	ver, err := version.GetVersion()
	if err != nil {
		panic(fmt.Errorf("fail to get version: %v", err))
	}

	app := &cli.App{
		Name:        serviceName,
		Version:     ver,
		Description: description,
		Usage:       usageText,
		Before: func(ctx *cli.Context) error {
			// TODO: elegent set or get env
			config.SetENV(config.ENVInfo{
				Proxy:        proxyAddress,
				SyncInterval: syncInterval,
				Contract:     contract,
				LogDir:       logDir,
				LogLevel:     logLevel,
			})
			return nil
		},
		Flags: []cli.Flag{
			// proxy address
			&cli.StringFlag{
				Name:        "proxy",
				Aliases:     []string{"p"},
				Usage:       "address of sphinx proxy",
				EnvVars:     []string{"ENV_PROXY"},
				Required:    true,
				Value:       "",
				Destination: &proxyAddress,
			},
			// sync interval
			&cli.Int64Flag{
				Name:        "sync-interval",
				Aliases:     []string{"si"},
				Usage:       "interval seconds of sync transaction on chain status",
				EnvVars:     []string{"ENV_SYNC_INTERVAL"},
				Required:    true,
				Value:       0,
				Destination: &syncInterval,
			},
			// contract id
			&cli.StringFlag{
				Name:        "contract",
				Aliases:     []string{"c"},
				Usage:       "id of contract",
				EnvVars:     []string{"ENV_CONTRACT"},
				Value:       "",
				Destination: &contract,
			},
			// log level
			&cli.StringFlag{
				Name:        "level",
				Aliases:     []string{"L"},
				Usage:       "level support debug|info|warning|error",
				EnvVars:     []string{"ENV_LOG_LEVEL"},
				Value:       "debug",
				DefaultText: "debug",
				Destination: &logLevel,
			},
			// log path
			&cli.StringFlag{
				Name:        "log",
				Aliases:     []string{"l"},
				Usage:       "log dir",
				EnvVars:     []string{"ENV_LOG_DIR"},
				Value:       "/var/log",
				DefaultText: "/var/log",
				Destination: &logDir,
			},
		},
		Commands: commands,
	}

	err = logger.Init(logger.DebugLevel, filepath.Join(config.GetENV().LogDir, "sphinx-plugin.log"))
	if err != nil {
		panic(fmt.Errorf("fail to init logger: %v", err))
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Sugar().Errorf("fail to run %v: %v", serviceName, err)
	}
}

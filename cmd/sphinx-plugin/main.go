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
	"golang.org/x/xerrors"
)

const (
	serviceName = "Sphinx Plugin"
	usageText   = "Sphinx Plugin Service"
)

var (
	proxyAddress = ""
	contractID   = ""
	logPath      = ""
	logLevel     = ""
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
		panic(xerrors.Errorf("Fail to get version: %v", err))
	}

	app := &cli.App{
		Name:        serviceName,
		Version:     ver,
		Description: description,
		Usage:       usageText,
		Before: func(ctx *cli.Context) error {
			config.SetENV(config.ENVInfo{
				Proxy:      proxyAddress,
				ContractID: contractID,
				LogPath:    logPath,
				LogLevel:   logLevel,
			})
			return nil
		},
		Flags: []cli.Flag{
			// proxy address
			&cli.StringFlag{
				Name:        "proxy",
				Aliases:     []string{"p"},
				Usage:       "address of sphinx proxy",
				EnvVars:     []string{"PROXY"},
				Required:    true,
				Value:       "",
				Destination: &proxyAddress,
			},
			// contract id
			&cli.StringFlag{
				Name:        "contract",
				Aliases:     []string{"c"},
				Usage:       "id of contract",
				EnvVars:     []string{"CONTRACT"},
				Value:       "",
				Destination: &contractID,
			},
			// log level
			&cli.StringFlag{
				Name:        "level",
				Aliases:     []string{"L"},
				Usage:       "level support debug|info|warning|error",
				EnvVars:     []string{"LEVEL"},
				Value:       "debug",
				DefaultText: "debug",
				Destination: &logLevel,
			},
			// log path
			&cli.StringFlag{
				Name:        "log",
				Aliases:     []string{"l"},
				Usage:       "log path",
				EnvVars:     []string{"LOG"},
				Value:       "/var/log",
				DefaultText: "/var/log",
				Destination: &logPath,
			},
		},
		Commands: commands,
	}

	err = logger.Init(logger.DebugLevel, filepath.Join(config.GetENV().LogPath, "sphinx-plugin.log"))
	if err != nil {
		panic(xerrors.Errorf("Fail to init logger: %v", err))
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Sugar().Errorf("fail to run %v: %v", serviceName, err)
	}
}

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/version"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	banner "github.com/common-nighthawk/go-figure"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

const (
	serviceName = "Sphinx Plugin"
	usageText   = "Sphinx Plugin Service"
)

func main() {
	commands := cli.Commands{
		runCmd,
	}

	description := fmt.Sprintf("my %v service cli\nFor help on any individual command run <%v COMMAND -h>\n",
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
		ArgsUsage:   "",
		Usage:       usageText,
		Flags:       nil,
		Commands:    commands,
	}

	err = config.Init("./", strings.ReplaceAll(serviceName, " ", ""))
	if err != nil {
		panic(xerrors.Errorf("Fail to create configuration: %v", err))
	}

	err = logger.Init(logger.DebugLevel, fmt.Sprintf("%v/%v.log", config.GetString(config.KeyLogDir), strings.ReplaceAll(serviceName, " ", "")))
	if err != nil {
		panic(xerrors.Errorf("Fail to init logger: %v", err))
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Sugar().Errorf("fail to run %v: %v", serviceName, err)
	}
}

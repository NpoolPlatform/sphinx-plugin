package main

import (
	"fmt"
	"log"
	"os"

	"github.com/NpoolPlatform/go-service-framework/pkg/version"
	"github.com/NpoolPlatform/sphinx-plugin/cmd/usdt"
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
	wanIP        string
	position     string
)

func main() {
	commands := cli.Commands{runCmd}
	commands = append(commands, usdt.USDTCmd...)
	description := fmt.Sprintf(
		"%v service cli\nFor help on any individual command run <%v COMMAND -h>\n",
		serviceName,
		serviceName,
	)

	banner.NewColorFigure(serviceName, "", "green", true).Print()
	vsion, err := version.GetVersion()
	if err != nil {
		panic(fmt.Errorf("fail to get version: %v", err))
	}

	app := &cli.App{
		Name:        serviceName,
		Version:     vsion,
		Description: description,
		Usage:       usageText,
		Commands:    commands,
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatalf("fail to run %v: %v", serviceName, err)
	}
}

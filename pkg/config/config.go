package config

import (
	"fmt"

	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

const (
	KeyLogDir          = "logdir"
	KeyAppID           = "appid"
	KeyHTTPPort        = "http_port"
	KeyGRPCPort        = "grpc_port"
	KeyPrometheusPort  = "prometheus_port"
	KeySphinxProxyAddr = "sphinx_proxy_addr"
	KeyContractID      = "contract_id"
	rootConfig         = "config"
)

func Init(configPath, appName string) error {
	viper.SetConfigName(fmt.Sprintf("%s.viper", appName))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath("./")
	viper.AddConfigPath(fmt.Sprintf("/etc/%v", appName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%v", appName))
	viper.AddConfigPath(".")

	// Following're must for every service
	// config:
	//   hostname: my-service.npool.top
	//   http_port: 32759
	//   grpc_port: 32789
	//   prometheus_port: 32799
	//   appid: "89089012783789789719823798127398",
	//   logdir: "/var/log"
	//
	if err := viper.ReadInConfig(); err != nil {
		return xerrors.Errorf("fail to init config: %v", err)
	}

	appID := viper.GetStringMap(rootConfig)[KeyAppID].(string)   //nolint
	logDir := viper.GetStringMap(rootConfig)[KeyLogDir].(string) //nolint

	fmt.Printf("appid: %v\n", appID)
	fmt.Printf("logdir: %v\n", logDir)
	return nil
}

func GetString(key string) string {
	return viper.GetStringMap(rootConfig)[key].(string)
}

func GetInt(key string) int {
	return viper.GetStringMap(rootConfig)[key].(int)
}

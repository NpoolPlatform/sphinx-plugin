package config

import (
	"fmt"

	"github.com/spf13/viper"
	"golang.org/x/xerrors"
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

	fmt.Printf("appid: %v\n", viper.GetString("appid"))
	fmt.Printf("logdir: %v\n", viper.GetString("logdir"))
	return nil
}

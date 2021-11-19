package env

import (
	"os"
)

const (
	ENVCOINAPI   = "ENV_COIN_API"
	ENVCOINTOKEN = "ENV_COIN_TOKEN"
)

func LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

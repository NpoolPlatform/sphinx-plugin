package eth

import "github.com/NpoolPlatform/sphinx-plugin/pkg/config"

// USDTContract ...
var USDTContract = func(chainet int64) string {
	switch chainet {
	case 1:
		return "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	case 1337:
		return config.GetString(config.KeyContract)
	}
	return ""
}

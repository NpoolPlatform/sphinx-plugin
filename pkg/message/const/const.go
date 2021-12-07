package constant

import (
	"strings"
	"time"
)

const (
	ServiceName = "sphinx-plugin.npool.top"
	GrpcTimeout = time.Second * 10
)

func FormatServiceName() string {
	return strings.Title(ServiceName)
}

package constant

import "strings"

const (
	ServiceName = "sphinx-plugin.npool.top"
)

func FormatServiceName() string {
	return strings.Title(ServiceName)
}

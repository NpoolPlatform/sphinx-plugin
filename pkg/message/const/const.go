package constant

import (
	"time"
)

const (
	ServiceName       = "sphinx-plugin.npool.top"
	GrpcTimeout       = time.Second * 10
	WaitMsgOutTimeout = time.Second * 40
)

type CtxKey int

const PluginSN CtxKey = iota

package constant

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
)

const (
	ServiceName       = "sphinx-plugin.npool.top"
	GrpcTimeout       = time.Second * 10
	WaitMsgOutTimeout = time.Second * 40
)

func SetPluginSN(ctx context.Context, pluginSN string) context.Context {
	md := metadata.New(map[string]string{"_pluginsn": pluginSN})
	return metadata.NewOutgoingContext(ctx, md)
}

func GetPluginSN(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if v, ok := md["_pluginsn"]; ok {
			return strings.Join(v, "-")
		}
	}
	return "pluginSN not set"
}

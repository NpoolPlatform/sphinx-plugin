package api

import (
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"google.golang.org/grpc"
)

type Server struct {
	sphinxplugin.UnimplementedPluginServer
}

func Register(server grpc.ServiceRegistrar) {
	sphinxplugin.RegisterPluginServer(server, &Server{})
}

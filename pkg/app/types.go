package app

import (
	"github.com/BrobridgeOrg/gravity-data-handler/pkg/eventbus"
	"github.com/BrobridgeOrg/gravity-data-handler/pkg/grpc_server"
	"github.com/BrobridgeOrg/gravity-data-handler/pkg/mux_manager"
)

type App interface {
	GetGRPCServer() grpc_server.Server
	GetMuxManager() mux_manager.Manager
	GetEventBus() eventbus.EventBus
}

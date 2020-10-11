package instance

import (
	"time"

	eventbus "github.com/BrobridgeOrg/gravity-data-handler/pkg/eventbus/service"
	grpc_server "github.com/BrobridgeOrg/gravity-data-handler/pkg/grpc_server/server"
	mux_manager "github.com/BrobridgeOrg/gravity-data-handler/pkg/mux_manager/manager"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AppInstance struct {
	done       chan bool
	muxManager *mux_manager.MuxManager
	grpcServer *grpc_server.Server
	eventBus   *eventbus.EventBus
}

func NewAppInstance() *AppInstance {

	a := &AppInstance{
		done: make(chan bool),
	}

	return a
}

func (a *AppInstance) Init() error {

	log.Info("Starting application")

	// Initializing modules
	a.muxManager = mux_manager.NewMuxManager(a)
	a.grpcServer = grpc_server.NewServer(a)
	a.eventBus = eventbus.NewEventBus(
		a,
		viper.GetString("nats.host"),
		eventbus.EventBusHandler{
			Reconnect: func(natsConn *nats.Conn) {
				log.Warn("re-connected to event server")
				//a.eventBus.InitSubscription()
			},
			Disconnect: func(natsConn *nats.Conn) {
				log.Error("event server was disconnected")
			},
		},
		eventbus.Options{
			PingInterval:        time.Duration(viper.GetInt64("nats.pingInterval")),
			MaxPingsOutstanding: viper.GetInt("nats.maxPingsOutstanding"),
			MaxReconnects:       viper.GetInt("nats.maxReconnects"),
		},
	)

	a.initMuxManager()

	// Initializing EventBus
	err := a.initEventBus()
	if err != nil {
		return err
	}

	// Initializing GRPC server
	err = a.initGRPCServer()
	if err != nil {
		return err
	}

	return nil
}

func (a *AppInstance) Uninit() {
}

func (a *AppInstance) Run() error {

	// GRPC
	go func() {
		err := a.runGRPCServer()
		if err != nil {
			log.Error(err)
		}
	}()

	err := a.runMuxManager()
	if err != nil {
		return err
	}

	<-a.done

	return nil
}

package app

import (
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	app "gravity-data-handler/app/interface"
	data_handler "gravity-data-handler/services/data_handler"

	pb "github.com/BrobridgeOrg/gravity-api/service/data_handler"
)

func (a *App) InitGRPCServer(host string) error {

	// Start to listen on port
	lis, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.WithFields(log.Fields{
		"host": host,
	}).Info("Starting gRPC server on " + host)

	// Create gRPC server
	s := grpc.NewServer()

	// Register data source adapter service
	dataHandlerService := data_handler.CreateService(app.AppImpl(a))
	pb.RegisterDataHandlerServer(s, dataHandlerService)
	reflection.Register(s)

	log.WithFields(log.Fields{
		"service": "DataHandler",
	}).Info("Registered service")

	// Starting server
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

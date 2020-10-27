package data_handler

import (
	"io"

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	app "github.com/BrobridgeOrg/gravity-data-handler/pkg/app"

	pb "github.com/BrobridgeOrg/gravity-api/service/data_handler"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var PushSuccess = pb.PushReply{
	Success: true,
}

type Service struct {
	app      app.App
	handler  *Handler
	incoming chan *pb.PushRequest
}

func NewService(a app.App) *Service {

	// Initialize handler and Load rule config
	handler := CreateHandler(a)
	err := handler.LoadRuleFile("./rules/rules.json")
	if err != nil {
		log.Error(err)
		return nil
	}

	// Preparing service
	service := &Service{
		app:      a,
		handler:  handler,
		incoming: make(chan *pb.PushRequest, 204800),
	}

	go service.eventHandler()

	return service
}

func (service *Service) Push(ctx context.Context, in *pb.PushRequest) (*pb.PushReply, error) {
	/*
		log.WithFields(log.Fields{
			"event": in.EventName,
		}).Info("Received event")
	*/

	// Handle event
	err := service.push(in.EventName, in.Payload)
	if err != nil {
		return &pb.PushReply{
			Success: false,
			Reason:  err.Error(),
		}, nil
	}

	return &PushSuccess, nil
}

func (service *Service) PushStream(stream pb.DataHandler_PushStreamServer) error {

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}
		/*
			id := atomic.AddUint64((*uint64)(&counter), 1)

			if id%1000 == 0 {
				log.Info(id)
			}
		*/
		service.pushAsync(in)
	}
}

// internal implementation
func (service *Service) eventHandler() {

	for {
		select {
		case req := <-service.incoming:
			err := service.push(req.EventName, req.Payload)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (service *Service) push(eventName string, payload []byte) error {

	// Handle event
	err := service.handler.ProcessEvent(eventName, payload)
	if err != nil {
		return err
	}
	return nil
}

func (service *Service) pushAsync(in *pb.PushRequest) error {
	service.incoming <- in
	return nil
}

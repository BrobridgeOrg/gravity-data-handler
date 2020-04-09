package data_handler

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	app "gravity-data-handler/app/interface"
	pb "gravity-data-handler/pb"
)

type Service struct {
	app     app.AppImpl
	handler *Handler
}

func CreateService(a app.AppImpl) *Service {

	// Initialize handler and Load rule config
	handler := CreateHandler(a)
	err := handler.LoadRuleFile("./rules/rules.json")
	if err != nil {
		log.Error(err)
		return nil
	}

	// Preparing service
	service := &Service{
		app:     a,
		handler: handler,
	}

	return service
}

func (service *Service) Push(ctx context.Context, in *pb.PushRequest) (*pb.PushReply, error) {

	log.WithFields(log.Fields{
		"event": in.EventName,
	}).Info("Received event")

	// Parse payload
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(in.Payload), &payload)
	if err != nil {
		return &pb.PushReply{
			Success: false,
			Reason:  err.Error(),
		}, nil
	}

	// Handle event
	err = service.handler.HandleEvent(in.EventName, payload)
	if err != nil {
		return &pb.PushReply{
			Success: false,
			Reason:  err.Error(),
		}, nil
	}

	return &pb.PushReply{
		Success: true,
	}, nil
}

package data_handler

import (
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
	app     app.App
	handler *Handler
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
		app:     a,
		handler: handler,
	}

	return service
}

func (service *Service) Push(ctx context.Context, in *pb.PushRequest) (*pb.PushReply, error) {
	/*
		log.WithFields(log.Fields{
			"event": in.EventName,
		}).Info("Received event")
	*/

	// Parse payload
	var payload map[string]interface{}
	err := json.Unmarshal(in.Payload, &payload)
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

	return &PushSuccess, nil
}

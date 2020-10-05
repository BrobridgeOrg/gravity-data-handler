package data_handler

import (
	"encoding/json"
	"errors"
	"fmt"
	app "gravity-data-handler/app/interface"
	"gravity-data-handler/services/data_handler/pipeline"
	"reflect"
	"time"

	pb "github.com/BrobridgeOrg/gravity-api/service/pipeline"
	"github.com/golang/protobuf/proto"

	log "github.com/sirupsen/logrus"
)

type Handler struct {
	app        app.AppImpl
	ruleConfig *RuleConfig
	pipeline   *pipeline.Manager
}

type Field struct {
	Name    string      `json:"name"`
	Value   interface{} `json:"value"`
	Primary bool        `json:"primary"`
}

type Projection struct {
	EventName  string  `json:"event"`
	Collection string  `json:"collection"`
	Method     string  `json:"method"`
	Fields     []Field `json:"fields"`
}

func CreateHandler(a app.AppImpl) *Handler {

	// Initialize pipelines
	opts := pipeline.NewOptions()
	opts.Handler = func(pipelineID int32, data interface{}) error {

		channel := fmt.Sprintf("gravity.pipeline.%d", pipelineID)
		//log.Info(channel)
		eb := a.GetEventBus()
		//err := eb.Emit("gravity.store.eventStored", data.([]byte))
		resp, err := eb.GetConnection().Request(channel, data.([]byte), time.Second*5)
		if err != nil {
			return err
		}

		var reply pb.PipelineReply
		err = proto.Unmarshal(resp.Data, &reply)
		if err != nil {
			return err
		}

		if !reply.Success {
			return errors.New(reply.Reason)
		}

		return nil
	}

	return &Handler{
		app:      a,
		pipeline: pipeline.NewManager(opts),
	}
}

func (handler *Handler) LoadRuleFile(filename string) error {

	// Load rule config
	config, err := LoadRuleFile("./rules/rules.json")
	if err != nil {
		log.Error(err)
		return nil
	}

	handler.ruleConfig = config

	return nil
}

func (handler *Handler) getPrimaryValueAsString(data interface{}) string {

	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.String:
		return data.(string)
	default:
		return fmt.Sprintf("%v", data)
	}
}

func (handler *Handler) HandleEvent(eventName string, payload map[string]interface{}) error {

	for _, rule := range handler.ruleConfig.Rules {

		if rule.Event != eventName {
			continue
		}

		// Preparing projection
		projection := Projection{
			EventName:  eventName,
			Collection: rule.Collection,
			Method:     rule.Method,
		}

		var primaryKey string

		for _, mapping := range rule.Mapping {

			if val, ok := payload[mapping.Source]; ok {

				if mapping.Primary {
					primaryKey = handler.getPrimaryValueAsString(val)
				}

				field := Field{
					Name:    mapping.Target,
					Value:   val,
					Primary: mapping.Primary,
				}

				projection.Fields = append(projection.Fields, field)
			}
		}

		// Publish to event store
		data, err := json.Marshal(&projection)
		if err != nil {
			return err
		}

		if primaryKey == "" {
			handler.pipeline.Dispatch(data)
		} else {
			handler.pipeline.Push(primaryKey, data)
		}
	}

	return nil
}

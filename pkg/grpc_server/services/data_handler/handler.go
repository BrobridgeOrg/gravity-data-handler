package data_handler

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	app "github.com/BrobridgeOrg/gravity-data-handler/pkg/app"

	pb "github.com/BrobridgeOrg/gravity-api/service/pipeline"
	"github.com/BrobridgeOrg/gravity-data-handler/pkg/grpc_server/services/data_handler/pipeline"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

var counter uint64

type Handler struct {
	app        app.App
	ruleConfig *RuleConfig
	pipeline   *pipeline.Manager
	channels   map[int32]string
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

var projectionPool = sync.Pool{
	New: func() interface{} {
		return &Projection{}
	},
}

var replyPool = sync.Pool{
	New: func() interface{} {
		return &pb.PipelineReply{}
	},
}

func CreateHandler(a app.App) *Handler {

	viper.SetDefault("pipeline.size", 256)
	pipelineSize := viper.GetInt32("pipeline.size")

	channels := make(map[int32]string)
	for i := int32(0); i <= pipelineSize; i++ {
		channels[i] = fmt.Sprintf("gravity.pipeline.%d", i)
	}

	// Initialize pipelines
	opts := pipeline.NewOptions()
	opts.Caps = pipelineSize
	opts.Handler = func(pipelineID int32, data interface{}) error {

		channel := channels[pipelineID]

		eb := a.GetEventBus()
		//err := eb.Emit("gravity.store.eventStored", data.([]byte))
		resp, err := eb.GetConnection().Request(channel, data.([]byte), time.Second*5)
		if err != nil {
			return err
		}

		// Parsing response
		reply := replyPool.Get().(*pb.PipelineReply)

		//		var reply pb.PipelineReply
		err = proto.Unmarshal(resp.Data, reply)
		if err != nil {
			// Release
			replyPool.Put(reply)
			return err
		}

		if !reply.Success {
			// Release
			replyPool.Put(reply)
			return errors.New(reply.Reason)
		}

		// Release
		replyPool.Put(reply)

		return nil
	}

	return &Handler{
		app:      a,
		pipeline: pipeline.NewManager(opts),
		channels: channels,
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
		projection := projectionPool.Get().(*Projection)
		projection.EventName = eventName
		projection.Method = rule.Method
		projection.Collection = rule.Collection
		projection.Fields = make([]Field, 0, len(rule.Mapping))

		/*
			projection := Projection{
				EventName:  eventName,
				Collection: rule.Collection,
				Method:     rule.Method,
			}
		*/
		var primaryKey string

		for _, mapping := range rule.Mapping {

			val, ok := payload[mapping.Source]
			if !ok {
				continue
			}

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

		// Publish to event store
		data, err := json.Marshal(&projection)
		projectionPool.Put(projection)
		if err != nil {
			return err
		}

		handler.pipeline.Push(primaryKey, data)
	}
	/*
		id := atomic.AddUint64((*uint64)(&counter), 1)
		if id%1000 == 0 {
			log.Info(id)
		}
	*/
	return nil
}

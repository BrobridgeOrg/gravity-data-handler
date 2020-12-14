package data_handler

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	app "github.com/BrobridgeOrg/gravity-data-handler/pkg/app"
	"github.com/lithammer/go-jump-consistent-hash"

	pb "github.com/BrobridgeOrg/gravity-api/service/pipeline"
	"github.com/cfsghost/gosharding"
	parallel_chunked_flow "github.com/cfsghost/parallel-chunked-flow"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

var counter uint64

type Handler struct {
	app        app.App
	ruleConfig *RuleConfig
	channels   map[int32]string
	shard      *gosharding.Shard
	preprocess *parallel_chunked_flow.ParallelChunkedFlow
}

type Field struct {
	Name    string      `json:"name"`
	Value   interface{} `json:"value"`
	Primary bool        `json:"primary,omitempty"`
}

type Projection struct {
	EventName  string  `json:"event"`
	Collection string  `json:"collection"`
	Method     string  `json:"method"`
	PrimaryKey string  `json:"primaryKey"`
	Fields     []Field `json:"fields"`
}

type Payload map[string]interface{}

type RawData struct {
	EventName string
	Payload   []byte
}

type Event struct {
	PrimaryKey string
	PipelineID int32
	Payload    Payload
	Rule       *Rule
}

type Request struct {
	PipelineID int32
	Payload    []byte
}

var rawDataPool = sync.Pool{
	New: func() interface{} {
		return &RawData{}
	},
}

var eventPool = sync.Pool{
	New: func() interface{} {
		return &Event{}
	},
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{}
	},
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
	//	viper.SetDefault("pipeline.workerCount", 32)
	pipelineSize := viper.GetInt32("pipeline.size")
	//	workerCount := viper.GetInt32("pipeline.workerCount")

	channels := make(map[int32]string)
	for i := int32(0); i <= pipelineSize; i++ {
		channels[i] = fmt.Sprintf("gravity.pipeline.%d", i)
	}

	handler := &Handler{
		app:      a,
		channels: channels,
	}

	// Initializing shard
	handler.initializeShard()

	// Initialize parapllel chunked flow
	pcfOpts := &parallel_chunked_flow.Options{
		BufferSize: 1024000,
		ChunkSize:  512,
		ChunkCount: 512,
		Handler: func(data interface{}, output chan interface{}) {

			rawData := data.(*RawData)

			// Parse payload
			var payload Payload
			err := json.Unmarshal(rawData.Payload, &payload)
			if err != nil {
				return
			}

			eventName := rawData.EventName
			rawDataPool.Put(rawData)

			for _, rule := range handler.ruleConfig.Rules {

				if rule.Event != eventName {
					continue
				}

				// Getting primary key
				primaryKey := handler.findPrimaryKey(rule, payload)

				event := eventPool.Get().(*Event)
				event.PrimaryKey = primaryKey
				event.PipelineID = jump.HashString(primaryKey, pipelineSize, jump.NewCRC64())
				event.Payload = payload
				event.Rule = rule

				output <- event
			}
		},
	}

	handler.preprocess = parallel_chunked_flow.NewParallelChunkedFlow(pcfOpts)

	go handler.EventReceiver()

	return handler
}

func (handler *Handler) EventReceiver() {
	for {
		select {
		case event := <-handler.preprocess.Output():
			// Push event to pipeline
			handler.shard.PushKV(event.(*Event).PrimaryKey, event)
		}
	}
}

func (handler *Handler) initializeShard() error {

	viper.SetDefault("pipeline.workerCount", 32)
	workerCount := viper.GetInt32("pipeline.workerCount")

	// Initializing shard
	options := gosharding.NewOptions()
	options.PipelineCount = workerCount
	options.BufferSize = 10240
	options.PrepareHandler = func(id int32, data interface{}, c chan interface{}) {

		for {
			data, err := handler.PreparePipelineData(id, data.(*Event))
			if err == nil {
				c <- data
				return
			}

			time.Sleep(time.Second)
		}
	}
	options.Handler = func(id int32, data interface{}) {

		for {
			err := handler.ProcessPipelineData(id, data.(*Request))
			if err == nil {
				return
			}

			time.Sleep(time.Second)
		}
	}

	// Create shard with options
	handler.shard = gosharding.NewShard(options)

	return nil
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

func (handler *Handler) findPrimaryKey(rule *Rule, payload Payload) string {

	if rule.PrimaryKey != "" {

		val, ok := payload[rule.PrimaryKey]
		if !ok {
			return ""
		}

		return handler.getPrimaryValueAsString(val)
	}
	/*
		for _, mapping := range rule.Mapping {

			val, ok := payload[mapping.Source]
			if !ok {
				continue
			}

			if mapping.Primary {
				return handler.getPrimaryValueAsString(val)
			}
		}
	*/
	return ""
}

func (handler *Handler) ProcessEvent(eventName string, data []byte) error {

	/*
		id := atomic.AddUint64((*uint64)(&counter), 1)
		if id%1000 == 0 {
			log.Info(id)
		}
	*/

	rawData := rawDataPool.Get().(*RawData)
	rawData.EventName = eventName
	rawData.Payload = data
	handler.preprocess.Push(rawData)

	return nil
}

func (handler *Handler) preparePacket(event *Event) []byte {

	// Preparing projection
	projection := projectionPool.Get().(*Projection)
	projection.EventName = event.Rule.Event
	projection.Method = event.Rule.Method
	projection.Collection = event.Rule.Collection
	projection.PrimaryKey = event.Rule.PrimaryKey
	projection.Fields = make([]Field, 0, len(event.Rule.Mapping))

	// pass throuh
	if len(event.Rule.Mapping) == 0 {
		for key, value := range event.Payload {

			field := Field{
				Name:  key,
				Value: value,
			}

			projection.Fields = append(projection.Fields, field)

		}
	} else {
		for _, mapping := range event.Rule.Mapping {

			// Getting value from payload
			val, ok := event.Payload[mapping.Source]
			if !ok {
				continue
			}

			field := Field{
				Name:  mapping.Target,
				Value: val,
				//			Primary: mapping.Primary,
			}

			projection.Fields = append(projection.Fields, field)
		}
	}

	// Convert to packet
	data, _ := json.Marshal(&projection)
	projectionPool.Put(projection)

	return data
}

func (handler *Handler) PreparePipelineData(workerID int32, event *Event) (interface{}, error) {

	request := requestPool.Get().(*Request)
	request.PipelineID = event.PipelineID
	request.Payload = handler.preparePacket(event)
	eventPool.Put(event)

	return request, nil
}

func (handler *Handler) ProcessPipelineData(workerID int32, request *Request) error {

	// Getting channel name
	channel := handler.channels[request.PipelineID]

	// Send request
	eb := handler.app.GetEventBus()
	resp, err := eb.GetConnection().Request(channel, request.Payload, time.Second*5)
	requestPool.Put(request)
	if err != nil {
		return err
	}

	// Parsing response
	reply := replyPool.Get().(*pb.PipelineReply)
	err = proto.Unmarshal(resp.Data, reply)
	if err != nil {
		// Release
		replyPool.Put(reply)
		return err
	}

	if !reply.Success {
		err = errors.New(reply.Reason)

		// Release
		replyPool.Put(reply)

		return err
	}

	// Release
	replyPool.Put(reply)

	return nil
}

package data_handler

import (
	"encoding/json"
	app "gravity-data-handler/app/interface"

	log "github.com/sirupsen/logrus"
)

type Handler struct {
	app        app.AppImpl
	ruleConfig *RuleConfig
}

type Field struct {
	Name    string      `json:"name"`
	Value   interface{} `json:"value"`
	Primary bool        `json:"primary"`
}

type Projection struct {
	EventName string  `json:"event"`
	Table     string  `json:"table"`
	Method    string  `json:"method"`
	Fields    []Field `json:"fields"`
}

func CreateHandler(a app.AppImpl) *Handler {

	return &Handler{
		app: a,
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

func (handler *Handler) HandleEvent(eventName string, payload map[string]interface{}) error {

	eb := handler.app.GetEventBus()

	for _, rule := range handler.ruleConfig.Rules {

		if rule.Event != eventName {
			continue
		}

		// Preparing projection
		projection := Projection{
			EventName: eventName,
			Table:     rule.Table,
			Method:    rule.Method,
		}

		for _, mapping := range rule.Mapping {

			if val, ok := payload[mapping.Source]; ok {

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

		err = eb.Emit("gravity.store.eventStored", data)
		if err != nil {
			return err
		}

	}

	return nil
}

package instance

import "github.com/BrobridgeOrg/gravity-data-handler/pkg/eventbus"

func (a *AppInstance) initEventBus() error {
	err := a.eventBus.Connect()
	if err != nil {
		return err
	}

	return nil
}

func (a *AppInstance) GetEventBus() eventbus.EventBus {
	return eventbus.EventBus(a.eventBus)
}

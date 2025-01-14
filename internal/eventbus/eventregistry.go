package eventbus

import "errors"

type EventRegistry struct {
	events map[string][]*Event
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		events: make(map[string][]*Event),
	}
}

func (r *EventRegistry) Register(events []*Event) error {
	if len(events) == 0 {
		return errors.New("no events to register")
	}
	for _, event := range events {
		if event == nil {
			return errors.New("cannot register a nil event")
		}
		r.events[event.Name] = append(r.events[event.Name], event)
	}
	return nil
}

func (r *EventRegistry) Import(registry *EventRegistry) error {
	if registry == nil {
		return errors.New("cannot import from a nil registry")
	}
	for name, events := range registry.events {
		r.events[name] = append(r.events[name], events...)
	}
	return nil
}

func (r *EventRegistry) Get(name string) ([]*Event, error) {
	eventList, exists := r.events[name]
	if !exists || len(eventList) == 0 {
		return nil, errors.New("events not found")
	}
	return eventList, nil
}

package eventbus

import (
	"fmt"
	"sync"
)

type EventPayload struct {
	Name    string
	Payload interface{}
}

type EventBus struct {
	mutex         *sync.Mutex
	onceInit      *sync.Once
	stopChannel   chan struct{}
	requestQueue  chan *EventPayload
	responseQueue chan *EventPayload
	eventRegistry *EventRegistry
	errorCallback chan error
}

func NewEventBus() (*EventBus, error) {
	eventRegistry := NewEventRegistry()
	eventBus := EventBus{
		mutex:         &sync.Mutex{},
		onceInit:      &sync.Once{},
		stopChannel:   make(chan struct{}),
		requestQueue:  make(chan *EventPayload, 100),
		responseQueue: make(chan *EventPayload, 100),
		eventRegistry: eventRegistry,
		errorCallback: make(chan error, 100),
	}
	return &eventBus, nil
}

func (eb *EventBus) Stop() {
	eb.onceInit.Do(func() {
		close(eb.stopChannel)
	})
}

func (eb *EventBus) Start() *EventBus {
	eb.onceInit.Do(func() {
		go func() {
			for {
				select {
				case <-eb.stopChannel:
					return
				case eventPayload := <-eb.requestQueue:
					eb.responseQueue <- eventPayload
				case eventPayload := <-eb.responseQueue:
					eb.ProcessEvent(eventPayload.Name, eventPayload.Payload)
				case err := <-eb.errorCallback:
					fmt.Printf("Error processing event: %v\n", err)
				}
			}
		}()
	})
	return eb
}

func (eb *EventBus) Publish(name string, payload interface{}) error {
	eb.requestQueue <- &EventPayload{
		Name:    name,
		Payload: payload,
	}
	return nil
}

func (eb *EventBus) ProcessEvent(eventName string, payload interface{}) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	defer func() {
		if r := recover(); r != nil {
			eb.errorCallback <- fmt.Errorf("Recovered from panic: %v", r)
		}
	}()

	events, err := eb.eventRegistry.Get(eventName)
	if err != nil {
		eb.errorCallback <- err
	}

	for _, event := range events {
		go func(e *Event) {
			output, err := event.handler(payload)
			if err != nil {
				if event.saga != nil {
					eb.requestQueue <- &EventPayload{
						Name:    *event.saga,
						Payload: output,
					}
				} else {
					eb.errorCallback <- err
				}
			}
			if event.next != nil {
				eb.requestQueue <- &EventPayload{
					Name:    event.next.name,
					Payload: output,
				}
			}
		}(event)
	}
}

func (eb *EventBus) Register(events []*Event) {
	eb.eventRegistry.Register(events)
}

func (eb *EventBus) Import(registry *EventRegistry) {
	eb.eventRegistry.Import(registry)
}

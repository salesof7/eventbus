package eventbus

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	tracer        trace.Tracer
}

func NewEventBus() (*EventBus, error) {
	tracer := otel.Tracer("eventbus")
	eventRegistry := NewEventRegistry()
	eventBus := EventBus{
		mutex:         &sync.Mutex{},
		onceInit:      &sync.Once{},
		stopChannel:   make(chan struct{}),
		requestQueue:  make(chan *EventPayload, 100),
		responseQueue: make(chan *EventPayload, 100),
		eventRegistry: eventRegistry,
		errorCallback: make(chan error, 100),
		tracer:        tracer,
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
					_, span := eb.tracer.Start(context.Background(), "ErrorCallback")
					span.SetAttributes(attribute.String("error", err.Error()))
					span.End()
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
	ctx, span := eb.tracer.Start(context.Background(), "ProcessEvent")
	defer span.End()

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
			_, eventSpan := eb.tracer.Start(ctx, "EventHandler", trace.WithAttributes(attribute.String("event_name", event.Name)))
			defer eventSpan.End()

			output, err := event.Handler(payload)
			if err != nil {
				if event.Saga != nil {
					eb.requestQueue <- &EventPayload{
						Name:    *event.Saga,
						Payload: output,
					}
				} else {
					eb.errorCallback <- err
				}
			}
			if event.Next != nil {
				eb.requestQueue <- &EventPayload{
					Name:    event.Next.Name,
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

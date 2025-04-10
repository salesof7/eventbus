package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type EventBroker interface {
	Publish(eventPayload *EventPayload, topic string) error
	Consume(responseQueue chan *EventPayload, errorCallback chan error) (*EventPayload, error)
}

type EventPayload struct {
	Name    string
	Payload interface{}
}

type EventBusConfig struct {
	RequestQueueSize  int
	ResponseQueueSize int
	ErrorQueueSize    int
	WorkerPoolSize    int
	BatchSize         int
	Timeout           time.Duration
}

type EventBus struct {
	mutex          *sync.Mutex
	onceStart      *sync.Once
	onceStop       *sync.Once
	stopChannel    chan struct{}
	requestQueue   chan *EventPayload
	responseQueue  chan *EventPayload
	eventRegistry  *EventRegistry
	errorCallback  chan error
	tracer         trace.Tracer
	meter          metric.Meter
	eventBroker    EventBroker
	workerPool     chan struct{}
	batch          []*EventPayload
	config         EventBusConfig
	eventCache     map[string][]*Event
	publishCounter metric.Int64Counter
	processCounter metric.Int64Counter
	publishLatency metric.Float64Histogram
	processLatency metric.Float64Histogram
	queueSize      metric.Int64Gauge
	errorCounter   metric.Int64Counter
}

func NewEventBus(eventBroker EventBroker, tracer trace.Tracer, config EventBusConfig) (*EventBus, error) {
	if tracer == nil {
		tracer = otel.Tracer("noop")
	}
	if config.RequestQueueSize == 0 {
		config.RequestQueueSize = 100
	}
	if config.ResponseQueueSize == 0 {
		config.ResponseQueueSize = 100
	}
	if config.ErrorQueueSize == 0 {
		config.ErrorQueueSize = 100
	}
	if config.WorkerPoolSize == 0 {
		config.WorkerPoolSize = 10
	}
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	eventRegistry := NewEventRegistry()
	meter := otel.Meter("eventbus")

	eventBus := EventBus{
		mutex:         &sync.Mutex{},
		onceStart:     &sync.Once{},
		onceStop:      &sync.Once{},
		stopChannel:   make(chan struct{}),
		requestQueue:  make(chan *EventPayload, config.RequestQueueSize),
		responseQueue: make(chan *EventPayload, config.ResponseQueueSize),
		errorCallback: make(chan error, config.ErrorQueueSize),
		eventRegistry: eventRegistry,
		eventBroker:   eventBroker,
		tracer:        tracer,
		meter:         meter,
		workerPool:    make(chan struct{}, config.WorkerPoolSize),
		batch:         make([]*EventPayload, 0, config.BatchSize),
		config:        config,
		eventCache:    make(map[string][]*Event),
	}

	var err error
	eventBus.publishCounter, err = meter.Int64Counter("eventbus.publish.count", metric.WithDescription("Number of events published"))
	if err != nil {
		return nil, err
	}
	eventBus.processCounter, err = meter.Int64Counter("eventbus.process.count", metric.WithDescription("Number of events processed"))
	if err != nil {
		return nil, err
	}
	eventBus.publishLatency, err = meter.Float64Histogram("eventbus.publish.latency", metric.WithDescription("Latency of publishing events"))
	if err != nil {
		return nil, err
	}
	eventBus.processLatency, err = meter.Float64Histogram("eventbus.process.latency", metric.WithDescription("Latency of processing events"))
	if err != nil {
		return nil, err
	}
	eventBus.queueSize, err = meter.Int64Gauge("eventbus.queue.size", metric.WithDescription("Current size of the queues"))
	if err != nil {
		return nil, err
	}
	eventBus.errorCounter, err = meter.Int64Counter("eventbus.errors", metric.WithDescription("Number of errors"))
	if err != nil {
		return nil, err
	}

	return &eventBus, nil
}

func (eb *EventBus) Stop() {
	eb.onceStop.Do(func() {
		close(eb.stopChannel)
	})
}

func (eb *EventBus) Start() *EventBus {
	eb.onceStart.Do(func() {

		go func() {
			for {
				select {
				case <-eb.stopChannel:
					return
				case eventPayload := <-eb.requestQueue:
					eb.batch = append(eb.batch, eventPayload)
					if len(eb.batch) >= eb.config.BatchSize {
						eb.publishBatch()
					}
				case eventPayload := <-eb.responseQueue:
					eb.ProcessEvent(eventPayload.Name, eventPayload.Payload)
				case err := <-eb.errorCallback:
					_, span := eb.tracer.Start(context.Background(), "errorCallback")
					span.SetAttributes(attribute.String("error", err.Error()))
					span.End()
					fmt.Printf("Error processing event: %v\n", err)
					eb.errorCounter.Add(context.Background(), 1, metric.WithAttributes(attribute.String("type", "callback")))
				default:
					if eb.eventBroker != nil {
						eb.eventBroker.Consume(eb.responseQueue, eb.errorCallback)
					}
				}
			}
		}()

		go func() {
			for {
				select {
				case <-eb.stopChannel:
					return
				default:
					eb.queueSize.Record(context.Background(), int64(len(eb.requestQueue)), metric.WithAttributes(attribute.String("queue", "request")))
					eb.queueSize.Record(context.Background(), int64(len(eb.responseQueue)), metric.WithAttributes(attribute.String("queue", "response")))
					eb.queueSize.Record(context.Background(), int64(len(eb.errorCallback)), metric.WithAttributes(attribute.String("queue", "error")))
					time.Sleep(1 * time.Second)
				}
			}
		}()
	})
	return eb
}

func (eb *EventBus) publishBatch() {
	ctx, cancel := context.WithTimeout(context.Background(), eb.config.Timeout)
	defer cancel()
	_, span := eb.tracer.Start(ctx, "PublishBatch")
	span.SetAttributes(attribute.Int("batch_size", len(eb.batch)))
	defer span.End()

	start := time.Now()
	if eb.eventBroker != nil {
		for _, event := range eb.batch {
			err := eb.eventBroker.Publish(event, "event_topic")
			if err != nil {

				eb.errorCallback <- fmt.Errorf("failed to publish message: %w", err)
				eb.errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("type", "publish")))
				span.RecordError(err)
			} else {
				eb.publishCounter.Add(ctx, 1)
			}
		}
	} else {
		for _, event := range eb.batch {
			eb.responseQueue <- event
			eb.publishCounter.Add(ctx, 1)
		}
	}
	duration := time.Since(start).Seconds()
	eb.publishLatency.Record(ctx, duration)
	eb.batch = eb.batch[:0]
}

func (eb *EventBus) Publish(name string, payload interface{}) error {
	select {
	case eb.requestQueue <- &EventPayload{Name: name, Payload: payload}:
		return nil
	default:
		eb.errorCounter.Add(context.Background(), 1, metric.WithAttributes(attribute.String("type", "queue_full")))
		return fmt.Errorf("request queue full")
	}
}

func (eb *EventBus) ProcessEvent(eventName string, payload interface{}) {
	ctx, span := eb.tracer.Start(context.Background(), "ProcessEvent")
	span.SetAttributes(attribute.String("event_name", eventName))
	defer span.End()

	eb.mutex.Lock()
	events, ok := eb.eventCache[eventName]
	if !ok {
		var err error
		events, err = eb.eventRegistry.Get(eventName)
		if err != nil {

			eb.errorCallback <- err
			eb.mutex.Unlock()
			return
		}
		eb.eventCache[eventName] = events
	}
	eb.mutex.Unlock()

	defer func() {
		if r := recover(); r != nil {
			eb.errorCallback <- fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	for _, event := range events {

		eb.workerPool <- struct{}{}
		go func(e *Event) {
			defer func() { <-eb.workerPool }()

			ctx, cancel := context.WithTimeout(ctx, eb.config.Timeout)
			defer cancel()

			_, eventSpan := eb.tracer.Start(ctx, "EventHandler", trace.WithAttributes(attribute.String("event_name", e.Name)))
			defer eventSpan.End()

			start := time.Now()
			eventSpan.AddEvent("Starting event processing")
			output, err := e.Handler(payload)
			eventSpan.AddEvent("Finished event processing")
			duration := time.Since(start).Seconds()
			eb.processLatency.Record(ctx, duration)
			eb.processCounter.Add(ctx, 1)

			if err != nil {
				eventSpan.RecordError(err)
				eventSpan.SetStatus(codes.Error, err.Error())
				eb.errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("type", "handler")))

				if e.Saga != nil {
					eb.requestQueue <- &EventPayload{Name: *e.Saga, Payload: output}
				} else {
					eb.errorCallback <- err
				}
			}
			if e.Next != nil {
				eb.requestQueue <- &EventPayload{Name: e.Next.Name, Payload: output}
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

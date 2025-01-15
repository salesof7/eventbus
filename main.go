package main

import (
	"fmt"
	"log"

	"github.com/salesof7/eventbus/internal/eventbus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("failed to create stdout exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("eventbus")
	bus, err := eventbus.NewEventBus(nil, tracer)
	if err != nil {
		log.Fatalf("failed to create eventbus: %v", err)
	}
	event := &eventbus.Event{
		Name: "test",
		Handler: func(payload interface{}) (interface{}, error) {
			fmt.Printf("Evento test recebido com payload: %v\n", payload)
			return nil, nil
		},
	}

	bus.Register([]*eventbus.Event{event})

	bus.Start()
	payload := struct {
		Name string
	}{
		Name: "Test Event",
	}

	err = bus.Publish("test", payload)
	if err != nil {
		log.Fatalf("failed to publish event: %v", err)
	}

	select {}
}

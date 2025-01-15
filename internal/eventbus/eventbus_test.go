package eventbus

import (
	"fmt"
	"testing"
	"time"
)

func TestNewEventBus(t *testing.T) {
	eb, err := NewEventBus(nil, nil)
	if err != nil {
		t.Fatalf("Error creating EventBus: %v", err)
	}
	if eb == nil {
		t.Fatal("Expected a non-nil EventBus")
	}
}

func TestEventBus_StartStop(t *testing.T) {
	eb, _ := NewEventBus(nil, nil)
	eb.Start()

	select {
	case <-time.After(100 * time.Millisecond):
	}

	eb.Stop()
	select {
	case <-time.After(100 * time.Millisecond):
	}
}

func TestEventBus_Publish(t *testing.T) {
	eb, _ := NewEventBus(nil, nil)
	eb.Start()

	err := eb.Publish("TestEvent", "Test Payload")
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	select {
	case eventPayload := <-eb.responseQueue:
		if eventPayload.Name != "TestEvent" {
			t.Fatalf("Expected event name 'TestEvent', got %s", eventPayload.Name)
		}
		if eventPayload.Payload != "Test Payload" {
			t.Fatalf("Expected payload 'Test Payload', got %v", eventPayload.Payload)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event was not processed in time")
	}
}

func TestEventBus_ProcessEvent(t *testing.T) {
	eventHandler := func(payload interface{}) (interface{}, error) {
		return "Processed " + payload.(string), nil
	}
	event := &Event{
		Name:    "TestEvent",
		Handler: eventHandler,
	}

	eb, _ := NewEventBus(nil, nil)
	eb.Start()
	eb.Register([]*Event{event})

	err := eb.Publish("TestEvent", "TestPayload")
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	select {
	case eventPayload := <-eb.responseQueue:
		if eventPayload.Name != "TestEvent" {
			t.Fatalf("Expected event name 'TestEvent', got %s", eventPayload.Name)
		}
		if eventPayload.Payload != "TestPayload" {
			t.Fatalf("Expected payload 'TestPayload', got %v", eventPayload.Payload)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event was not processed in time")
	}
}

func TestEventBus_ProcessEventWithError(t *testing.T) {
	eventHandler := func(payload interface{}) (interface{}, error) {
		return nil, fmt.Errorf("handler error")
	}

	event := &Event{
		Name:    "TestEventWithError",
		Handler: eventHandler,
	}

	eb, _ := NewEventBus(nil, nil)
	eb.Start()
	eb.Register([]*Event{event})

	err := eb.Publish("TestEventWithError", "TestPayload")
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	select {
	case eventPayload := <-eb.responseQueue:
		if eventPayload.Name != "TestEventWithError" {
			t.Fatalf("Expected event name 'TestEventWithError', got %s", eventPayload.Name)
		}
		t.Fatal("Expected an error during event processing, but none occurred")
	case <-time.After(100 * time.Millisecond):
	}
}

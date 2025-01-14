package eventbus

import (
	"testing"
)

func TestNewEventRegistry(t *testing.T) {
	registry := NewEventRegistry()
	if registry == nil {
		t.Fatal("Expected NewEventRegistry to return a non-nil value")
	}
	if len(registry.events) != 0 {
		t.Fatal("Expected new registry to have no events")
	}
}

func TestEventRegistry_Register(t *testing.T) {
	registry := NewEventRegistry()

	event1 := &Event{Name: "Event1"}
	event2 := &Event{Name: "Event2"}

	err := registry.Register([]*Event{event1, event2})
	if err != nil {
		t.Fatalf("Unexpected error during registration: %v", err)
	}

	if len(registry.events) != 2 {
		t.Fatalf("Expected 2 events to be registered, got %d", len(registry.events))
	}

	err = registry.Register([]*Event{event1})
	if err != nil {
		t.Fatalf("Unexpected error during duplicate registration: %v", err)
	}

	if len(registry.events["Event1"]) != 2 {
		t.Fatalf("Expected 2 instances of Event1 to be registered, got %d", len(registry.events["Event1"]))
	}
}

func TestEventRegistry_Import(t *testing.T) {
	registry := NewEventRegistry()
	otherRegistry := NewEventRegistry()

	event1 := &Event{Name: "Event1"}
	event2 := &Event{Name: "Event2"}
	event3 := &Event{Name: "Event1"}

	err := otherRegistry.Register([]*Event{event1, event2, event3})
	if err != nil {
		t.Fatalf("Unexpected error during registration in other registry: %v", err)
	}

	err = registry.Import(otherRegistry)
	if err != nil {
		t.Fatalf("Unexpected error during import: %v", err)
	}

	if len(registry.events) != 2 {
		t.Fatalf("Expected 2 unique event names to be imported, got %d", len(registry.events))
	}

	if len(registry.events["Event1"]) != 2 {
		t.Fatalf("Expected 2 instances of Event1 after import, got %d", len(registry.events["Event1"]))
	}

	if len(registry.events["Event2"]) != 1 {
		t.Fatalf("Expected 1 instance of Event2 after import, got %d", len(registry.events["Event2"]))
	}
}

func TestEventRegistry_Get(t *testing.T) {
	registry := NewEventRegistry()

	event1 := &Event{Name: "Event1"}
	event2 := &Event{Name: "Event2"}

	err := registry.Register([]*Event{event1, event2})
	if err != nil {
		t.Fatalf("Unexpected error during registration: %v", err)
	}

	events, err := registry.Get("Event1")
	if err != nil {
		t.Fatalf("Expected to find Event1, got error: %v", err)
	}
	if len(events) != 1 || events[0] != event1 {
		t.Fatalf("Expected to retrieve Event1, got %v", events)
	}

	_, err = registry.Get("Event3")
	expectedError := "events not found"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("Expected error '%s', but got '%v'", expectedError, err)
	}
}

package eventbus

import (
	"testing"
)

func TestEventFlow_AddEvent(t *testing.T) {
	baseEvent := &Event{name: "BaseEvent"}
	flow := &EventFlow{baseEvent: baseEvent, lastEvent: baseEvent}

	event1 := &Event{name: "Event1"}
	flow.AddEvent(event1)

	if flow.lastEvent != event1 {
		t.Errorf("Expected last event to be %v, got %v", event1, flow.lastEvent)
	}
	if flow.baseEvent.next != event1 {
		t.Errorf("Expected baseEvent.next to point to %v, got %v", event1, flow.baseEvent.next)
	}
}

func TestEventFlow_Next(t *testing.T) {
	baseEvent := &Event{name: "BaseEvent"}
	flow := &EventFlow{baseEvent: baseEvent, lastEvent: baseEvent}

	event1 := &Event{name: "Event1"}
	flow.Next(event1)

	if flow.lastEvent != event1 {
		t.Errorf("Expected last event to be %v, got %v", event1, flow.lastEvent)
	}

	if flow.baseEvent.next != event1 {
		t.Errorf("Expected baseEvent.next to point to %v, got %v", event1, flow.baseEvent.next)
	}

	event2 := &Event{name: "Event2"}
	flow.Next(event2)

	if flow.lastEvent != event2 {
		t.Errorf("Expected last event to be %v, got %v", event2, flow.lastEvent)
	}
	if event1.next != event2 {
		t.Errorf("Expected event1.next to point to %v, got %v", event2, event1.next)
	}
}

func TestEventFlow_Saga(t *testing.T) {
	baseEvent := &Event{name: "BaseEvent"}
	flow := &EventFlow{baseEvent: baseEvent, lastEvent: baseEvent}

	saga := &Event{name: "SagaEvent"}
	flow.Saga(saga)

	if flow.lastSaga != saga {
		t.Errorf("Expected last saga to be %v, got %v", saga, flow.lastSaga)
	}

	if baseEvent.saga == nil || *baseEvent.saga != saga.name {
		t.Errorf("Expected baseEvent.saga to point to %v, got %v", saga.name, baseEvent.saga)
	}

	saga2 := &Event{name: "SagaEvent2"}
	flow.Saga(saga2)

	if saga2.next != saga {
		t.Errorf("Expected saga2.next to point to %v, got %v", saga, saga2.next)
	}

	if flow.lastSaga != saga2 {
		t.Errorf("Expected last saga to be %v, got %v", saga2, flow.lastSaga)
	}
}

func TestEventFlow_Flat(t *testing.T) {
	baseEvent := &Event{name: "BaseEvent"}
	flow := &EventFlow{baseEvent: baseEvent, lastEvent: baseEvent}

	event1 := &Event{name: "Event1"}
	event2 := &Event{name: "Event2"}
	saga := &Event{name: "SagaEvent"}

	flow.Next(event1).Next(event2).Saga(saga)

	events := flow.Flat()

	expectedNames := []string{"BaseEvent", "Event1", "Event2", "SagaEvent"}
	if len(events) != len(expectedNames) {
		t.Fatalf("Expected %d events, got %d", len(expectedNames), len(events))
	}

	for i, e := range events {
		if e.name != expectedNames[i] {
			t.Errorf("Expected event %d to be %v, got %v", i, expectedNames[i], e.name)
		}
	}
}

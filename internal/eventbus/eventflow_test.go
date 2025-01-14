package eventbus

import (
	"testing"
)

func TestEventFlow_AddEvent(t *testing.T) {
	baseEvent := &Event{Name: "BaseEvent"}
	flow := &EventFlow{baseEvent: baseEvent, lastEvent: baseEvent}

	event1 := &Event{Name: "Event1"}
	flow.AddEvent(event1)

	if flow.lastEvent != event1 {
		t.Errorf("Expected last event to be %v, got %v", event1, flow.lastEvent)
	}
	if flow.baseEvent.Next != event1 {
		t.Errorf("Expected baseEvent.Next to point to %v, got %v", event1, flow.baseEvent.Next)
	}
}

func TestEventFlow_Next(t *testing.T) {
	baseEvent := &Event{Name: "BaseEvent"}
	flow := &EventFlow{baseEvent: baseEvent, lastEvent: baseEvent}

	event1 := &Event{Name: "Event1"}
	flow.Next(event1)

	if flow.lastEvent != event1 {
		t.Errorf("Expected last event to be %v, got %v", event1, flow.lastEvent)
	}

	if flow.baseEvent.Next != event1 {
		t.Errorf("Expected baseEvent.Next to point to %v, got %v", event1, flow.baseEvent.Next)
	}

	event2 := &Event{Name: "Event2"}
	flow.Next(event2)

	if flow.lastEvent != event2 {
		t.Errorf("Expected last event to be %v, got %v", event2, flow.lastEvent)
	}
	if event1.Next != event2 {
		t.Errorf("Expected event1.Next to point to %v, got %v", event2, event1.Next)
	}
}

func TestEventFlow_Saga(t *testing.T) {
	baseEvent := &Event{Name: "BaseEvent"}
	flow := &EventFlow{baseEvent: baseEvent, lastEvent: baseEvent}

	saga := &Event{Name: "SagaEvent"}
	flow.Saga(saga)

	if flow.lastSaga != saga {
		t.Errorf("Expected last saga to be %v, got %v", saga, flow.lastSaga)
	}

	if baseEvent.Saga == nil || *baseEvent.Saga != saga.Name {
		t.Errorf("Expected baseEvent.saga to point to %v, got %v", saga.Name, baseEvent.Saga)
	}

	saga2 := &Event{Name: "SagaEvent2"}
	flow.Saga(saga2)

	if saga2.Next != saga {
		t.Errorf("Expected saga2.Next to point to %v, got %v", saga, saga2.Next)
	}

	if flow.lastSaga != saga2 {
		t.Errorf("Expected last saga to be %v, got %v", saga2, flow.lastSaga)
	}
}

func TestEventFlow_Flat(t *testing.T) {
	baseEvent := &Event{Name: "BaseEvent"}
	flow := &EventFlow{baseEvent: baseEvent, lastEvent: baseEvent}

	event1 := &Event{Name: "Event1"}
	event2 := &Event{Name: "Event2"}
	saga := &Event{Name: "SagaEvent"}

	flow.Next(event1).Next(event2).Saga(saga)

	events := flow.Flat()

	expectedNames := []string{"BaseEvent", "Event1", "Event2", "SagaEvent"}
	if len(events) != len(expectedNames) {
		t.Fatalf("Expected %d events, got %d", len(expectedNames), len(events))
	}

	for i, e := range events {
		if e.Name != expectedNames[i] {
			t.Errorf("Expected event %d to be %v, got %v", i, expectedNames[i], e.Name)
		}
	}
}

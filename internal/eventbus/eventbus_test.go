package eventbus

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockEventBroker struct{}

func (m *MockEventBroker) Publish(eventPayload *EventPayload, topic string) error {
	return nil
}

func (m *MockEventBroker) Consume(responseQueue chan *EventPayload, errorCallback chan error) (*EventPayload, error) {
	return nil, nil
}

func TestNewEventBus(t *testing.T) {
	config := EventBusConfig{
		RequestQueueSize:  10,
		ResponseQueueSize: 10,
		ErrorQueueSize:    10,
		WorkerPoolSize:    5,
		BatchSize:         5,
		Timeout:           2 * time.Second,
	}
	eventBus, err := NewEventBus(&MockEventBroker{}, nil, config)
	assert.NoError(t, err)
	assert.NotNil(t, eventBus)
	assert.Equal(t, 10, cap(eventBus.requestQueue), "Tamanho da fila de requisições deve ser 10")
	assert.Equal(t, 10, cap(eventBus.responseQueue), "Tamanho da fila de respostas deve ser 10")
	assert.Equal(t, 10, cap(eventBus.errorCallback), "Tamanho da fila de erros deve ser 10")
	assert.Equal(t, 5, cap(eventBus.workerPool), "Tamanho do worker pool deve ser 5")
	assert.Equal(t, 5, cap(eventBus.batch), "Tamanho do batch deve ser 5")
}

func TestRegisterEvents(t *testing.T) {
	eventBus, _ := NewEventBus(nil, nil, EventBusConfig{})
	event := &Event{
		Name: "test_event",
		Handler: func(payload interface{}) (interface{}, error) {
			return nil, nil
		},
	}
	eventBus.Register([]*Event{event})
	events, err := eventBus.eventRegistry.Get("test_event")
	assert.NoError(t, err)
	assert.Len(t, events, 1, "Deve haver exatamente 1 evento registrado")
	assert.Equal(t, "test_event", events[0].Name, "O nome do evento deve ser 'test_event'")
}

func TestPublishAndProcessEvent(t *testing.T) {
	eventBus, _ := NewEventBus(nil, nil, EventBusConfig{
		BatchSize: 1,
	})
	var handlerCalled bool
	event := &Event{
		Name: "test_event",
		Handler: func(payload interface{}) (interface{}, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	eventBus.Register([]*Event{event})
	eventBus.Start()
	defer eventBus.Stop()

	err := eventBus.Publish("test_event", nil)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.True(t, handlerCalled, "O handler deve ser chamado após a publicação")
}

func TestErrorHandling(t *testing.T) {
	eventBus, _ := NewEventBus(nil, nil, EventBusConfig{
		BatchSize: 1,
	})

	testErrorChan := make(chan error, 1)
	eventBus.errorCallback = testErrorChan

	event := &Event{
		Name: "error_event",
		Handler: func(payload interface{}) (interface{}, error) {
			return nil, fmt.Errorf("handler error")
		},
	}
	eventBus.Register([]*Event{event})
	eventBus.Start()
	defer eventBus.Stop()

	err := eventBus.Publish("error_event", nil)
	assert.NoError(t, err)

	select {
	case err := <-testErrorChan:
		assert.EqualError(t, err, "handler error", "O erro retornado pelo handler deve ser capturado")
	case <-time.After(1 * time.Second):
		t.Error("Esperava um erro no canal errorCallback, mas não foi recebido")
	}
}

func TestQueueFull(t *testing.T) {
	eventBus, _ := NewEventBus(nil, nil, EventBusConfig{
		RequestQueueSize: 1,
	})
	eventBus.Start()
	defer eventBus.Stop()

	err := eventBus.Publish("test_event", nil)
	assert.NoError(t, err)

	err = eventBus.Publish("test_event", nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "request queue full", "Deve retornar erro quando a fila está cheia")
}

func TestStop(t *testing.T) {
	eventBus, _ := NewEventBus(nil, nil, EventBusConfig{})
	eventBus.Start()

	eventBus.Stop()

	select {
	case <-eventBus.stopChannel:

	default:
		t.Error("O canal stopChannel deveria estar fechado após Stop()")
	}
}

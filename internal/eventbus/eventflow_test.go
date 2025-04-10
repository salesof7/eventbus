package eventbus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEventFlow(t *testing.T) {
	ef := &EventFlow{}
	assert.Nil(t, ef.baseEvent, "baseEvent deve ser nil inicialmente")
	assert.Nil(t, ef.lastEvent, "lastEvent deve ser nil inicialmente")
	assert.Nil(t, ef.lastSaga, "lastSaga deve ser nil inicialmente")
}

func TestNextSingleEvent(t *testing.T) {
	ef := &EventFlow{}
	event := &Event{Name: "event1"}
	ef.Next(event)

	assert.Equal(t, event, ef.baseEvent, "baseEvent deve ser o evento adicionado")
	assert.Equal(t, event, ef.lastEvent, "lastEvent deve ser o evento adicionado")
	assert.Nil(t, ef.lastSaga, "lastSaga deve permanecer nil")
}

func TestNextMultipleEvents(t *testing.T) {
	ef := &EventFlow{}
	event1 := &Event{Name: "event1"}
	event2 := &Event{Name: "event2"}
	event3 := &Event{Name: "event3"}

	ef.Next(event1).Next(event2).Next(event3)

	assert.Equal(t, event1, ef.baseEvent, "baseEvent deve ser o primeiro evento")
	assert.Equal(t, event3, ef.lastEvent, "lastEvent deve ser o último evento")
	assert.Equal(t, event2, event1.Next, "event1.Next deve apontar para event2")
	assert.Equal(t, event3, event2.Next, "event2.Next deve apontar para event3")
	assert.Nil(t, event3.Next, "event3.Next deve ser nil")
	assert.Nil(t, ef.lastSaga, "lastSaga deve permanecer nil")
}

func TestSagaSingleSaga(t *testing.T) {
	ef := &EventFlow{}
	event := &Event{Name: "event1"}
	saga := &Event{Name: "saga1"}

	ef.Next(event).Saga(saga)

	assert.Equal(t, event, ef.baseEvent, "baseEvent deve ser o evento inicial")
	assert.Equal(t, event, ef.lastEvent, "lastEvent deve ser o evento inicial")
	assert.Equal(t, "saga1", *event.Saga, "event.Saga deve apontar para o nome do saga")
	assert.Equal(t, saga, ef.lastSaga, "lastSaga deve ser o saga adicionado")
	assert.Nil(t, saga.Next, "saga.Next deve ser nil")
}

func TestSagaMultipleSagas(t *testing.T) {
	ef := &EventFlow{}
	event1 := &Event{Name: "event1"}
	event2 := &Event{Name: "event2"}
	saga1 := &Event{Name: "saga1"}
	saga2 := &Event{Name: "saga2"}

	ef.Next(event1).Saga(saga1).Next(event2).Saga(saga2)

	assert.Equal(t, event1, ef.baseEvent, "baseEvent deve ser o primeiro evento")
	assert.Equal(t, event2, ef.lastEvent, "lastEvent deve ser o segundo evento")
	assert.Equal(t, "saga1", *event1.Saga, "event1.Saga deve apontar para saga1")
	assert.Equal(t, "saga2", *event2.Saga, "event2.Saga deve apontar para saga2")
	assert.Equal(t, saga2, ef.lastSaga, "lastSaga deve ser o segundo saga")
	assert.Equal(t, saga1, saga2.Next, "saga2.Next deve apontar para saga1")
	assert.Nil(t, saga1.Next, "saga1.Next deve ser nil")
}

func TestFlatEmpty(t *testing.T) {
	ef := &EventFlow{}
	events := ef.Flat()
	assert.Empty(t, events, "Flat deve retornar uma lista vazia para um EventFlow vazio")
}

func TestFlatWithEventsAndSagas(t *testing.T) {
	ef := &EventFlow{}
	event1 := &Event{Name: "event1"}
	event2 := &Event{Name: "event2"}
	saga1 := &Event{Name: "saga1"}
	saga2 := &Event{Name: "saga2"}

	ef.Next(event1).Saga(saga1).Next(event2).Saga(saga2)

	events := ef.Flat()
	assert.Len(t, events, 4, "Flat deve retornar todos os eventos e sagas")
	assert.Contains(t, events, event1, "Flat deve incluir event1")
	assert.Contains(t, events, event2, "Flat deve incluir event2")
	assert.Contains(t, events, saga1, "Flat deve incluir saga1")
	assert.Contains(t, events, saga2, "Flat deve incluir saga2")
}

func TestFlatWithCycle(t *testing.T) {
	ef := &EventFlow{}
	event1 := &Event{Name: "event1"}
	event2 := &Event{Name: "event2"}
	event1.Next = event2
	event2.Next = event1

	ef.Next(event1)
	events := ef.Flat()
	assert.Len(t, events, 2, "Flat deve lidar com ciclos e retornar apenas eventos únicos")
	assert.Contains(t, events, event1, "Flat deve incluir event1")
	assert.Contains(t, events, event2, "Flat deve incluir event2")
}

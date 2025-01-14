package eventbus

type EventFlow struct {
	baseEvent *Event
	lastEvent *Event
	lastSaga  *Event
}

func (ef *EventFlow) AddEvent(event *Event) *EventFlow {
	ef.lastEvent.next = event
	ef.lastEvent = event
	return ef
}

func (ef *EventFlow) Next(event *Event) *EventFlow {
	if ef.lastEvent == nil {
		return ef.AddEvent(event)
	} else {
		ef.lastEvent.next = event
		ef.lastEvent = event
	}
	return ef
}

func (ef *EventFlow) Saga(saga *Event) *EventFlow {
	if ef.lastEvent != nil {
		ef.lastEvent.saga = &saga.name
	}
	saga.next = nil
	if ef.lastSaga != nil {
		saga.next = ef.lastSaga
	}
	ef.lastSaga = saga
	return ef
}

func (ef *EventFlow) Flat() []*Event {
	var events []*Event
	visited := make(map[*Event]bool)

	var traverse func(e *Event)
	traverse = func(e *Event) {
		if e == nil || visited[e] {
			return
		}
		visited[e] = true
		events = append(events, e)
		traverse(e.next)
	}
	traverse(ef.baseEvent)
	traverse(ef.lastSaga)
	return events
}

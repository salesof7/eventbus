package eventbus

type EventFlow struct {
	baseEvent *Event
	lastEvent *Event
	lastSaga  *Event
}

func (ef *EventFlow) Next(event *Event) *EventFlow {
	if ef.lastEvent == nil {
		ef.baseEvent = event
		ef.lastEvent = event
	} else {
		ef.lastEvent.Next = event
		ef.lastEvent = event
	}
	return ef
}

func (ef *EventFlow) Saga(saga *Event) *EventFlow {
	if ef.lastEvent != nil {
		ef.lastEvent.Saga = &saga.Name
	}
	saga.Next = nil
	if ef.lastSaga != nil {
		saga.Next = ef.lastSaga
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
		traverse(e.Next)
	}
	traverse(ef.baseEvent)
	traverse(ef.lastSaga)
	return events
}
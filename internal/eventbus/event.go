package eventbus

type Event struct {
	Name    string
	Saga    *string
	Next    *Event
	Handler func(payload interface{}) (interface{}, error)
}

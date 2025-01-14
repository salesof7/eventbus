package eventbus

type Event struct {
	name    string
	saga    *string
	next    *Event
	handler func(payload interface{}) (interface{}, error)
}

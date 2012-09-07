package interfaces

type Event interface {
	Trigger(eventname string, params ...interface{})
	Iterate(eventname string, stopfunc interface{}, params ...interface{})
}

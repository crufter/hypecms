package interfaces

type Event interface {
	Trigger(eventname string, params ...interface{})
	Iterate(eventname string, stopfunc interface{}, params ...interface{})
}

type Caller interface {
	Call(string, string, string, interface{}, ...interface{})
	Names(string, string) []string
	Matches(string, string, string, interface{}) bool
	Has(string, string, string) bool
}

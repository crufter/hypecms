package interfaces

import "reflect"

type Event interface {
	Trigger(eventname string, params ...interface{})
	Iterate(eventname string, stopfunc interface{}, params ...interface{})
}

type Caller interface {
	Call(string, string, interface{}, ...interface{}) error
	Names(string) []string
	Matches(string, string, interface{}) bool
	Has(string, string) bool
	Inputs(string, string) []reflect.Type
	Outputs(string, string) []reflect.Type
}

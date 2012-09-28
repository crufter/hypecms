package interfaces

import (
	"reflect"
	"labix.org/v2/mgo/bson"
)

type Event interface {
	Trigger(eventname string, params ...interface{})
	Iterate(eventname string, stopfunc interface{}, params ...interface{})
}

type Caller interface {
	Call(string, string, interface{}, ...interface{}) error
	Names(string) []string
	Matches(string, string, interface{}) bool
	Has(string, string) bool
	Exists(string) bool
	Inputs(string, string) []reflect.Type
	Outputs(string, string) []reflect.Type
}

type Filter interface {
	Ids() ([]bson.ObjectId, error)
	Find() ([]interface{}, error)
	Insert(map[string]interface{}) error
	Update(map[string]interface{}) error
	Remove() error
}

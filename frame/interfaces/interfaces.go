package interfaces

import (
	"reflect"
	"labix.org/v2/mgo/bson"
)

type Event interface {
	Trigger(eventname string, params ...interface{})
	Iterate(eventname string, stopfunc interface{}, params ...interface{})
}

type Method interface {
	Call(interface{}, ...interface{}) error
	Matches(interface{}) bool
	InputTypes() []reflect.Type
	OutputTypes() []reflect.Type
}

type Instance interface {
	HasMethod(string) bool
	MethodNames() []string
	Method(string) Method
}

type Module interface {
	Instance() Instance
	Exists() bool
}

type Speaker interface {
	IsNoun(string) bool
	NounHasVerb(string, string) bool
}

type Filter interface {
	Ids() ([]bson.ObjectId, error)
	Find() ([]interface{}, error)
	Insert(map[string]interface{}) error
	Update(map[string]interface{}) error
	Remove() error
}

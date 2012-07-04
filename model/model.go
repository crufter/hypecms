package model

import "github.com/opesun/hypecms/api/context"

type Event interface{
	Param(params ...interface{})
	TriggerAll(eventnames ...string)
	Trigger(eventname string, params ...interface{})
}

// Temp hack.
var Convert = context.Convert
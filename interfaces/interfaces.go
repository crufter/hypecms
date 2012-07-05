package interfaces

type Event interface{
	Param(params ...interface{})
	TriggerAll(eventnames ...string)
	Trigger(eventname string, params ...interface{})
}

type View interface{
	Publish(key string, value ...interface{})
}
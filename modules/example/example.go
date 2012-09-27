// Package example implements a minimalistic plugin for hypeCMS.
package example

import (
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/routep"
	"labix.org/v2/mgo/bson"
)

// Create a type only to spare ourselves from typing map[string]interface{} every time.
type m map[string]interface{}

func (v *C) ExampleAnything(i int) error {
	v.uni.Dat["example_value"] = i
	return nil
}

func (h *C) Front() (bool, error) {
	var hijacked bool
	if _, err := routep.Comp("/example", h.uni.P); err == nil {
		hijacked = true                                    	// This stops the main front loop from executing any further modules.
	}
	return hijacked, nil
}

func (h *C) Install(id bson.ObjectId) error {
	example_options := m{
		"example": "any value",
	}
	q := m{"_id": id}
	upd := m{
		"$addToSet": m{
			"Hooks.Front": "example",
		},
		"$set": m{
			"Modules.example": example_options,
		},
	}
	return h.uni.Db.C("options").Update(q, upd)
}

func (h *C) Uninstall(id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$pull": m{
			"Hooks.Front": "example",
		},
		"$unset": m{
			"Modules.example": 1,
		},
	}
	return h.uni.Db.C("options").Update(q, upd)
}

type C struct {
	uni *context.Uni
}

func (c *C) Hooks(uni *context.Uni) {
	c.uni = uni
}
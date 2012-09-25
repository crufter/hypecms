// Package example implements a minimalistic plugin for hypeCMS.
package example

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/routep"
	"labix.org/v2/mgo/bson"
)

// Create a type only to spare ourselves from typing map[string]interface{} every time.
type m map[string]interface{}

func (v *V) ExampleAnything(i int) error {
	v.uni.Dat["example_value"] = i
	return nil
}

func (h *H) Front() (bool, error) {
	var hijacked bool
	if _, err := routep.Comp("/example", h.uni.P); err == nil {
		hijacked = true                                    	// This stops the main front loop from executing any further modules.
	}
	return hijacked, nil
}

func (h *H) Install(id bson.ObjectId) error {
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

func (h *H) Uninstall(id bson.ObjectId) error {
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

type H struct {
	uni *context.Uni
}

func Hooks(uni *context.Uni) *H {
	return &H{uni}
}

type V struct {
	uni *context.Uni
}

func Views(uni *context.Uni) *V {
	return &V{uni}
}

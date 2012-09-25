// Package skeleton implements a minimalistic but idiomatic plugin for hypeCMS.
package skeleton

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/routep"
	"labix.org/v2/mgo/bson"
)

// Create a type only to spare ourselves from typing map[string]interface{} every time.
type m map[string]interface{}

type V struct {
	uni *context.Uni
}

type H struct {
	uni *context.Uni
}

func Views(uni *context.Uni) *V {
	return &V{uni}
}

func Hooks(uni *context.Uni) *H {
	return &H{uni}
}

func (v *V) Front() (bool, error) {
	var hijacked bool
	if _, err := routep.Comp("/skeleton", v.uni.P); err == nil {
		hijacked = true                                    	// This stops the main front loop from executing any further modules.
	}
	return hijacked, nil
}

func (h *H) Install(id bson.ObjectId) error {
	skeleton_options := m{
		"example": "any value",
	}
	q := m{"_id": id}
	upd := m{
		"$addToSet": m{
			"Hooks.Front": "skeleton",
		},
		"$set": m{
			"Modules.skeleton": skeleton_options,
		},
	}
	return h.uni.Db.C("options").Update(q, upd)
}

func (h *H) Uninstall(id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$pull": m{
			"Hooks.Front": "skeleton",
		},
		"$unset": m{
			"Modules.skeleton": 1,
		},
	}
	return h.uni.Db.C("options").Update(q, upd)
}

// Package skeleton implements a minimalistic but idiomatic plugin for HypeCMS.
package skeleton

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"labix.org/v2/mgo/bson"
	"fmt"
)

// Create a type only to spare ourselves from typing map[string]interface{} every time.
type m map[string]interface{}

// mod.GetHook accesses certain functions dynamically trough this.
var Hooks = map[string]interface{} {
	"Front":     Front,
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}

// main.runFrontHooks invokes this trough mod.GetHook.
func Front(uni *context.Uni, hijacked *bool) error {
	if _, err := routep.Comp("/skeleton", uni.P); err == nil {
		*hijacked = true		// This is important, this stops the main front loop from executing any further modules.
		uni.Put("Hello, this is the skeleton module here.")		// This is just a basic output to allow you to see your freshly written module.
	}
	// You can insert code here which will decide wich view to call.
	return nil
}

// main.runBackHooks invokes this trough mod.GetHook.
func Back(uni *context.Uni, action string) error {
	switch action {
	// You can dispatch your background operations here.
	}
	return nil
}

// main.runDebug invokes this trough mod.GetHook.
func Test(uni *context.Uni) error {
	front := jsonp.HasVal(uni.Opt, "Hooks.Front", "skeleton")
	if !front {
		return fmt.Errorf("Not subscribed to front hook.")
	}
	return nil
}

// admin.AD invokes this trough mod.GetHook.
func AD(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"skeleton/index"}
	// You can dispatch your different admin views here, based on url structure.
	return nil
}

// admin.Install invokes this trough mod.GetHook.
func Install(uni *context.Uni, id bson.ObjectId) error {
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
	return uni.Db.C("options").Update(q, upd)
}

// Admin Install invokes this trough mod.GetHook.
func Uninstall(uni *context.Uni, id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$pull": m{
			"Hooks.Front": "skeleton",
		},
		"$unset": m{
			"Modules.skeleton": 1,
		},
	}
	return uni.Db.C("options").Update(q, upd)
}

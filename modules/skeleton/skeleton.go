// Package skeleton implements a minimalistic but idiomatic plugin for HypeCMS.
package skeleton

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"launchpad.net/mgo/bson"
	"fmt"
)

// Create a type only to spare ourselves from typing map[string]interface{} every time.
type m map[string]interface{}

// mod.GetHook accesses certain functions dynamically trough this.
var Hooks = map[string]func(*context.Uni) error {
	"Front":     Front,
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}

// main.runFrontHooks invokes this trough mod.GetHook.
func Front(uni *context.Uni) error {
	if _, err := routep.Comp("/skeleton", uni.P); err == nil {
		uni.Dat["_hijacked"] = true		// This is important, this stops the main front loop from executing any further modules.
		uni.Put("Hello, this is the skeleton module here.")		// This is just a basic output to allow you to see your freshly written module.
	}
	// You can insert code here which will decide wich view to call.
	return nil
}

// main.runBackHooks invokes this trough mod.GetHook.
func Back(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
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
func Install(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	skeleton_options := m{
		"example": "any value",
	}
	return uni.Db.C("options").Update(m{"_id": id}, m{"$addToSet": m{"Hooks.Front": "skeleton"}, "$set": m{"Modules.skeleton": skeleton_options}})
}

// Admin Install invokes this trough mod.GetHook.
func Uninstall(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	return uni.Db.C("options").Update(m{"_id": id}, m{"$pull": m{"Hooks.Front": "skeleton"}, "$unset": m{"Modules.skeleton": 1}})
}

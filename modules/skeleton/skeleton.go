// Package skeleton implements a minimalistic but idiomatic plugin for HypeCMS.
package skeleton

import(
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"launchpad.net/mgo/bson"
)

// Create a type only to spare ourselves from typing map[string]interface{} every time.
type m map[string]interface{}
var Hooks = map[string]func(*context.Uni){
	"Front":		Front,
	"Back":			Back,
	"Install": 		Install,
	"Uninstall": 	Uninstall,
	"Test":			Test,
	"AD":			AD,
}

func Front(uni *context.Uni) {
	if _, ok := routep.Comp("/skeleton", uni.P); ok == "" {
		uni.Dat["_hijacked"] = true
		uni.Put("Hello, this is the skeleton module here.")
	}
}

func Back(uni *context.Uni) {
}

func Test(uni *context.Uni) {
	res := map[string]interface{}{}
	res["Front"] = jsonp.HasVal(uni.Opt, "Hooks.Front", "skeleton")
	_, ok := jsonp.Get(uni.Opt, "Modules.Skeleton")
	res["Modules"] = ok
	uni.Dat["_cont"] = res
}

func AD(uni *context.Uni) {
	uni.Dat["_points"] = []string{"skeleton/index"}
}

func Install(uni *context.Uni) {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	skeleton_options := m{
		"example": "any value",
	}
	uni.Db.C("options").Update(m{"_id": id}, m{ "$addToSet": m{ "Hooks.Front": "skeleton"}, "$set": m{"Modules.skeleton": skeleton_options }})
}

func Uninstall(uni *context.Uni) {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	uni.Db.C("options").Update(m{"_id": id}, m{ "$pull": m{ "Hooks.Front": "skeleton"}, "$unset": m{"Modules.skeleton": 1 }})
}
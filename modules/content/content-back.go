package content

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/scut"
	"github.com/opesun/jsonp"
	"github.com/opesun/extract"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
)

type m map[string]interface{}

var Hooks = map[string]func(*context.Uni){
	"AD":        AD,
	"Front":     Front,
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
}

// Find slug value be given key.
func FindContent(db *mgo.Database, key, val string) (map[string]interface{}, bool) {
	query := make(bson.M)
	query[key] = val
	var v interface{}
	db.C("contents").Find(query).One(&v)
	if v == nil {
		return nil, false
	}
	return context.Convert(v).(map[string]interface{}), true
}

func Test(uni *context.Uni) {
	res := make(map[string]interface{})
	res["Front"] = jsonp.HasVal(uni.Opt, "Hooks.Front", "content")
	_, ok := jsonp.Get(uni.Opt, "Modules.Content")
	res["Modules"] = ok
	uni.Dat["_cont"] = res
}

func Install(uni *context.Uni) {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	content_options := m{
		"types": m {
			"blog": m{
				"rules" : m{
					"title": 1, "content": 1, "_created": 1,
				},
			},
		},
	}
	uni.Db.C("options").Update(m{"_id": id}, m{"$addToSet": m{"Hooks.Front": "content"}, "$set": m{"Modules.content": content_options}})
}

func Uninstall(uni *context.Uni) {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	uni.Db.C("options").Update(m{"_id": id}, m{"$pull": m{"Hooks.Front": "content"}, "$unset": m{"Modules.content": 1}})
}

func SaveTypeConfig(uni *context.Uni) {
	// id := scut.CreateOptCopy(uni.Db)
}

func SaveConfig(uni *context.Uni) {
	// id := scut.CreateOptCopy(uni.Db)
}

// TODO: Move Ins, Upd, Del to other package since they can be used with all modules similar to content.
// TODO: Separate the shared processes of Insert/Update (type and rule checking, extracting)
func Ins(uni *context.Uni) {
	res := map[string]interface{}{}
	typ, hastype := uni.Req.Form["type"]
	if !hastype {
		res["success"] = false; res["reason"] = "No type sent from form when inserting content."; uni.Dat["_cont"] = res; return
	}
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types." + typ[0] + ".rules")
	if !hasrule {
		res["success"] = false; res["reason"] = "Can't find content type " + typ[0]; uni.Dat["_cont"] = res; return
	}
	ins_dat, extr_err := extract.New(rule.(map[string]interface{})).ExtractForm(uni.Req.Form)
	ins_dat["type"] = typ[0]
	if extr_err != nil {
		res["success"] = false; res["reason"] = extr_err.Error(); uni.Dat["_cont"] = res; return
	}
	ins_err := scut.Inud(uni, ins_dat, &res, "contents", "insert", "")
	if ins_err != nil {
		res["success"] = false; res["reason"] = ins_err.Error(); uni.Dat["_cont"] = res; return
	}
	res["success"] = true
	uni.Dat["_cont"] = res
}

// TODO: Separate the shared processes of Insert/Update (type and rule checking, extracting)
func Upd(uni *context.Uni) {
	res := map[string]interface{}{}
	id, hasid := uni.Req.Form["id"]
	if !hasid {
		res["success"] = false; res["reason"] = "No id sent from form when updating content."; uni.Dat["_cont"] = res; return
	}
	typ, hastype := uni.Req.Form["type"]
	if !hastype {
		res["success"] = false; res["reason"] = "No type sent from form when updating content."; uni.Dat["_cont"] = res; return
	}
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types." + typ[0] + ".rules")
	if !hasrule {
		res["success"] = false; res["reason"] = "Can't find content type " + typ[0]; uni.Dat["_cont"] = res; return
	}
	upd_dat, extr_err := extract.New(rule.(map[string]interface{})).ExtractForm(uni.Req.Form)
	upd_dat["type"] = typ[0]
	if extr_err != nil {
		res["success"] = false; res["reason"] = extr_err.Error(); uni.Dat["_cont"] = res; return
	}
	upd_err := scut.Inud(uni, upd_dat, &res, "contents", "update", id[0])
	if upd_err != nil {
		res["success"] = false; res["reason"] = upd_err.Error(); uni.Dat["_cont"] = res; return
	}
	res["success"] = true
	uni.Dat["_cont"] = res
}

func Del(uni *context.Uni) {
	res := map[string]interface{}{}
	id, has := uni.Req.Form["id"]
	if !has {
		res["success"] = false; res["reason"] = "No id sent from form when deleting content."; uni.Dat["_cont"] = res; return
	}
	scut.Inud(uni, nil, &res, "contents", "delete", id[0])
	uni.Dat["_cont"] = res
}

func minLev(opt map[string]interface{}, op string) int {
	if v, ok := jsonp.Get(opt, "Modules.content." + op + "_level"); ok {
		return int(v.(float64))
	}
	return 300	// This is sparta.
}

func Back(uni *context.Uni) {
	action := uni.Dat["_action"].(string)
	_, ok := jsonp.Get(uni.Opt, "Modules.content")
	if !ok {
		uni.Put("No content options.")
		return
	}
	level := uni.Dat["_user"].(map[string]interface{})["level"].(int)
	if minLev(uni.Opt, action) > level {
		res := map[string]interface{}{}
		res["success"] = false
		res["reason"] = "You have no rights to do content action " + action
		uni.Dat["_cont"] = res
		return
	}
	switch action {
	case "insert":
		Ins(uni)
	case "update":
		Upd(uni)
	case "delete":
		Del(uni)
	case "save_config":
		SaveTypeConfig(uni)
	default:
		uni.Put("Can't find action named \"" + action + "\" in user module.")
	}
}

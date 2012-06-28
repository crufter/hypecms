package content

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/scut"
	"github.com/opesun/jsonp"
	"github.com/opesun/extract"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"fmt"
)

type m map[string]interface{}

var Hooks = map[string]func(*context.Uni) error {
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

func Test(uni *context.Uni) error {
	res := make(map[string]interface{})
	res["Front"] = jsonp.HasVal(uni.Opt, "Hooks.Front", "content")
	_, ok := jsonp.Get(uni.Opt, "Modules.Content")
	res["Modules"] = ok
	uni.Dat["_cont"] = res
	return nil
}

func Install(uni *context.Uni) error {
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
	return uni.Db.C("options").Update(m{"_id": id}, m{"$addToSet": m{"Hooks.Front": "content"}, "$set": m{"Modules.content": content_options}})
}

func Uninstall(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	return uni.Db.C("options").Update(m{"_id": id}, m{"$pull": m{"Hooks.Front": "content"}, "$unset": m{"Modules.content": 1}})
}

func SaveTypeConfig(uni *context.Uni) error {
	// id := scut.CreateOptCopy(uni.Db)
	return nil
}

func SaveConfig(uni *context.Uni) error {
	// id := scut.CreateOptCopy(uni.Db)
	return nil
}

// TODO: Move Ins, Upd, Del to other package since they can be used with all modules similar to content.
// TODO: Separate the shared processes of Insert/Update (type and rule checking, extracting)
func Ins(uni *context.Uni) error {
	typ, hastype := uni.Req.Form["type"]
	if !hastype {
		return fmt.Errorf("No type sent from form when inserting content.")
	}
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types." + typ[0] + ".rules")
	if !hasrule {
		return fmt.Errorf("Can't find content type " + typ[0])
	}
	ins_dat, extr_err := extract.New(rule.(map[string]interface{})).ExtractForm(uni.Req.Form)
	ins_dat["type"] = typ[0]
	if extr_err != nil {
		return extr_err
	}
	return scut.Inud(uni, ins_dat, "contents", "insert", "")
}

// TODO: Separate the shared processes of Insert/Update (type and rule checking, extracting)
func Upd(uni *context.Uni) error {
	id, hasid := uni.Req.Form["id"]
	if !hasid {
		return fmt.Errorf("No id sent from form when updating content.")
	}
	typ, hastype := uni.Req.Form["type"]
	if !hastype {
		return fmt.Errorf("No type sent from form when updating content.")
	}
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types." + typ[0] + ".rules")
	if !hasrule {
		return fmt.Errorf("Can't find content type " + typ[0])
	}
	upd_dat, extr_err := extract.New(rule.(map[string]interface{})).ExtractForm(uni.Req.Form)
	upd_dat["type"] = typ[0]
	if extr_err != nil {
		return extr_err
	}
	return scut.Inud(uni, upd_dat, "contents", "update", id[0])
}

func Del(uni *context.Uni) error {
	id, has := uni.Req.Form["id"]
	if !has {
		return fmt.Errorf("No id sent from form when deleting content.")
	}
	return scut.Inud(uni, nil, "contents", "delete", id[0])
}

func minLev(opt map[string]interface{}, op string) int {
	if v, ok := jsonp.Get(opt, "Modules.content." + op + "_level"); ok {
		return int(v.(float64))
	}
	return 300	// This is sparta.
}

func Back(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
	_, ok := jsonp.Get(uni.Opt, "Modules.content")
	if !ok {
		return fmt.Errorf("No content options.")
	}
	level := uni.Dat["_user"].(map[string]interface{})["level"].(int)
	if minLev(uni.Opt, action) > level {
		return fmt.Errorf("You have no rights to do content action " + action)
	}
	var r error
	switch action {
	case "insert":
		r = Ins(uni)
	case "update":
		r = Upd(uni)
	case "delete":
		r = Del(uni)
	case "save_config":
		r = SaveTypeConfig(uni)
	default:
		return fmt.Errorf("Can't find action named \"" + action + "\" in user module.")
	}
	return r 
}

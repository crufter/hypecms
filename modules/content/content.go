package content

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/scut"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"github.com/opesun/extract"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"encoding/json"
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

func Front(uni *context.Uni) {
	//uni.Put("article module runs")
	m, err := routep.Comp("/{slug}", uni.Req.URL.Path)
	if err == nil {
		content, found := FindContent(uni.Db, "slug", m["slug"])
		if found {
			uni.Dat["_hijacked"] = true
			uni.Dat["_points"] = []string{"content"}
			uni.Dat["content"] = content
		}
	}

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

func getSidebar(uni *context.Uni) []string {
	menu := []string{}
	types, has := jsonp.Get(uni.Opt, "Modules.content.types")
	if !has {
		panic("There are no content types.")
	}
	for i, _ := range types.(map[string]interface{}) {
		menu = append(menu, i)
	}
	return menu
}

func Index(uni *context.Uni) {
	uni.Dat["content_menu"] = getSidebar(uni)
	v := uni.Db.C("contents").Find(nil).Sort(m{"_created":-1})
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/index"}
}

func List(uni *context.Uni) {
	uni.Dat["content_menu"] = getSidebar(uni)
	ma, err := routep.Comp("/admin/content/list/{type}", uni.Req.URL.Path)
	if err != nil {
		uni.Put("Bad url at list."); return
	}
	typ, has := ma["type"]
	if !has {
		uni.Put("Can not extract typ at list."); return
	}
	v := uni.Db.C("contents").Find(m{"type":typ}).Sort(m{"_created":-1})
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/list"}
}

func TypeConfig(uni *context.Uni) {
	uni.Dat["content_menu"] = getSidebar(uni)
	ma, err := routep.Comp("/admin/content/type-config/{type}", uni.Req.URL.Path)
	if err != nil {
		uni.Put("Bad url at type config."); return
	}
	typ, has := ma["type"]
	if !has {
		uni.Put("Can not extract typ at type config."); return
	}
	op, ok := jsonp.Get(uni.Opt, "Modules.content.types." + typ)
	if !ok {
		uni.Put("Can not find content type " + typ + " in options."); return
	}
	uni.Dat["type"] = typ
	uni.Dat["type_options"], _ = json.MarshalIndent(op, "", "    ")
	uni.Dat["_points"] = []string{"content/type-config"}
}

func Config(uni *context.Uni) {
	uni.Dat["content_menu"] = getSidebar(uni)
	op, _ := jsonp.Get(uni.Opt, "Modules.content")
	v, err := json.MarshalIndent(op, "", "    ")
	if err != nil {
		uni.Put("Can't marshal content options.")
		return
	}
	uni.Dat["content_options"] = string(v)
	uni.Dat["_points"] = []string{"content/config"}
}

func Edit(uni *context.Uni) {
	uni.Dat["content_menu"] = getSidebar(uni)
	ma, err := routep.Comp("/admin/content/edit/{type}", uni.Req.URL.Path)
	if err != nil {
		uni.Put("Bad url at edit."); return
	}
	typ, hast := ma["type"]
	if !hast {
		uni.Put("Can't extract type at edit.")
		return
	}
	uni.Dat["content_type"] = typ
	rules, has := jsonp.Get(uni.Opt, "Modules.content.types." + typ + ".rules")
	if !has {
		uni.Put("Can't find rules of " + typ)
		return
	}
	uni.Dat["type"] = typ
	uni.Dat["rules"] = rules
	uni.Dat["_points"] = []string{"content/edit"}
}

func AD(uni *context.Uni) {
	m, _ := routep.Comp("/admin/content/{view}", uni.Req.URL.Path)
	switch m["view"] {
	case "":
		Index(uni)
	case "config":
		Config(uni)
	case "type-config":
		TypeConfig(uni)
	case "edit":
		Edit(uni)
	case "list":
		List(uni)
	}
}

func SaveTypeConfig(uni *context.Uni) {
	// id := scut.CreateOptCopy(uni.Db)
}

func SaveConfig(uni *context.Uni) {
	// id := scut.CreateOptCopy(uni.Db)
}

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
	dat, err := extract.New(rule.(map[string]interface{})).ExtractForm(uni.Req.Form)
	if err != nil {
		res["success"] = false
		res["reason"] = err.Error()
		uni.Dat["_cont"] = res
		return
	}
	id := uni.Req.Form["_id"][0]
	err = scut.Inud(uni, dat, &res, "contents", "update", id)
	uni.Dat["_cont"] = res
}

// TODO: Separate the shared processes of Insert/Update (type and rule checking, extracting)
func Upd(uni *context.Uni) {
	res := map[string]interface{}{}
	_, hasid := uni.Req.Form["_id"]
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
	dat, err := extract.New(rule.(map[string]interface{})).ExtractForm(uni.Req.Form)
	if err != nil {
		res["success"] = false
		res["reason"] = err.Error()
		uni.Dat["_cont"] = res
		return
	}
	id := uni.Req.Form["_id"][0]
	err = scut.Inud(uni, dat, &res, "contents", "update", id)
	uni.Dat["_cont"] = res
}

func Del(uni *context.Uni) {
	res := map[string]interface{}{}
	_, has := uni.Req.Form["_id"]
	if !has {
		res["success"] = false; res["reason"] = "No id sent from form when deleting content."; uni.Dat["_cont"] = res; return
	}
	id := uni.Req.Form["_id"][0]
	scut.Inud(uni, nil, &res, "contents", "delete", id)
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
	had_action := true
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
		had_action = false
	}
	if !had_action {
		uni.Put("Can't find action named \"" + action + "\" in user module.")
	}
}

package content

import(
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/routep"
	"github.com/opesun/jsonp"
	"launchpad.net/mgo/bson"
	"github.com/opesun/hypecms/api/scut"
	"encoding/json"
)

func Front(uni *context.Uni) error {
	ed, ed_err := routep.Comp("/admin/content/edit/{type}/{id}", uni.P)
	if ed_err == nil {
		ulev, hasu := jsonp.GetI(uni.Opt, "_user.level")
		if !hasu {
			panic("No user level found, or it is not an integer.")
			return nil
		}
		_, hasid := ed["id"]
		if hasid && ulev < minLev(uni.Opt, "edit") {
			panic("You have no rights to edit a content.")
			return nil
		} else if ulev < minLev(uni.Opt, "insert") {
			panic("You have no rights to insert a content.")
			return nil
		}
		uni.Dat["_hijacked"] = true
		Edit(uni, ed)
		uni.Dat["_points"] = []string{"edit-content"}	// Must contain require content/edit-form.t to work well.
		return nil
	}
	m, err := routep.Comp("/{slug}", uni.P)
	if err == nil {
		content, found := FindContent(uni.Db, "slug", m["slug"])
		if found {
			uni.Dat["_hijacked"] = true
			uni.Dat["_points"] = []string{"content"}
			uni.Dat["content"] = content
		}
	}
	return nil
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
	var v []interface{}
	uni.Db.C("contents").Find(nil).Sort(m{"_created":-1}).All(&v)
	scut.Strify(v)
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/index"}
}

func List(uni *context.Uni) {
	ma, err := routep.Comp("/admin/content/list/{type}", uni.Req.URL.Path)
	if err != nil {
		uni.Put("Bad url at list."); return
	}
	typ, has := ma["type"]
	if !has {
		uni.Put("Can not extract typ at list."); return
	}
	var v []interface{}
	uni.Db.C("contents").Find(m{"type":typ}).Sort(m{"_created":-1}).All(&v)
	scut.Strify(v)
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/list"}
}

func TypeConfig(uni *context.Uni) {
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
	op, _ := jsonp.Get(uni.Opt, "Modules.content")
	v, err := json.MarshalIndent(op, "", "    ")
	if err != nil {
		uni.Put("Can't marshal content options.")
		return
	}
	uni.Dat["content_options"] = string(v)
	uni.Dat["_points"] = []string{"content/config"}
}

// Called from both admin and outside editing.
func Edit(uni *context.Uni, ma map[string]string) {
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
	id, hasid := ma["id"]
	var indb interface{}
	if hasid {
		uni.Dat["op"] = "update"
		uni.Db.C("contents").Find(m{"_id": bson.ObjectIdHex(id)}).One(&indb)
		uni.Dat["content"] = indb
	} else {
		uni.Dat["op"] = "insert"
	}
	rs := []interface{}{}
	for i, v := range rules.(map[string]interface{}) {
		field := map[string]interface{}{"fieldname":i,"v":v}
		if indb != nil {
			field["value"] = indb.(bson.M)[i]
		}
		rs = append(rs, field)
	}
	uni.Dat["fields"] = rs
}

func AEdit(uni *context.Uni) {
	ma, err := routep.Comp("/admin/content/edit/{type}/{id}", uni.Req.URL.Path)
	if err != nil {
		uni.Put("Bad url at edit."); return
	}
	Edit(uni, ma)
	uni.Dat["_points"] = []string{"content/edit"}
}

func AD(uni *context.Uni) error {
	m, _ := routep.Comp("/admin/content/{view}", uni.Req.URL.Path)
	uni.Dat["content_menu"] = getSidebar(uni)
	switch m["view"] {
	case "":
		Index(uni)
	case "config":
		Config(uni)
	case "type-config":
		TypeConfig(uni)
	case "edit":
		AEdit(uni)
	case "list":
		List(uni)
	default:
		uni.Put("Unkown content view.")
	}
	return nil
}
package content

import(
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/routep"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/hypecms/api/scut"
	"github.com/opesun/hypecms/modules/content/model"
	"encoding/json"
	"fmt"
	"github.com/opesun/resolver"
	//"strings"
)

type m map[string]interface{}

func UserEdit(uni *context.Uni, urimap map[string]string) error {
	ulev, hasu := jsonp.GetI(uni.Dat, "_user.level")
	if !hasu {
		return fmt.Errorf("No user level found, or it is not an integer.")
	}
	_, hasid := urimap["id"]
	if hasid && ulev < minLev(uni.Opt, "edit") {
		return fmt.Errorf("You have no rights to edit a content.")
	} else if ulev < minLev(uni.Opt, "insert") {
		return fmt.Errorf("You have no rights to insert a content.")
	}
	Edit(uni, urimap)
	uni.Dat["_hijacked"] = true
	uni.Dat["_points"] = []string{"edit-content"}	// Must contain require content/edit-form.t to work well.
	return nil
}

func TagView(uni *context.Uni, urimap map[string]string) error {
	list, err := content_model.ListContentsByTag(uni.Db, urimap["slug"])
	if err != nil {
		uni.Dat["error"] = err.Error()
	} else {
		uni.Dat["content_list"] = list
	}
	uni.Dat["_hijacked"] = true
	uni.Dat["_points"] = []string{"tag"}
	return nil
}

func TagSearch(uni *context.Uni, urimap map[string]string) error {
	list, err := content_model.TagSearch(uni.Db, urimap["slug"])
	if err != nil {
		uni.Dat["error"] = err.Error()
	} else {
		uni.Dat["tag_list"] = list
	}
	uni.Dat["_hijacked"] = true
	uni.Dat["_points"] = []string{"tag-search"}
	return nil
}

func ContentView(uni *context.Uni, content_map map[string]string) error {
	types, ok := jsonp.Get(uni.Opt, "Modules.content.types")
	if !ok {
		return nil
	}
	slug_keymap := map[string]struct{}{}
	for _, v := range types.(map[string]interface{}) {
		type_conf := v.(map[string]interface{})
		if slugval, has := type_conf["slug"]; has {
			slug_keymap[slugval.(string)] = struct{}{}
		} else {
			slug_keymap["_id"] = struct{}{}
		}
	}
	slug_keys := []string{}
	for i, _ := range slug_keymap {
		slug_keys = append(slug_keys, i)
	}
	content, found := content_model.FindContent(uni.Db, slug_keys, content_map["slug"])
	if found {
		uni.Dat["_hijacked"] = true
		uni.Dat["_points"] = []string{"content"}
		uni.Dat["content"] = content
	}
	return nil
}

func Front(uni *context.Uni) error {
	edit_map, edit_err := routep.Comp("/content/edit/{type}/{id}", uni.P)
	if edit_err == nil {
		return UserEdit(uni, edit_map)
	}
	tag_map, tag_err := routep.Comp("/tag/{slug}", uni.P)
	// Tag view: list contents in that category.
	if tag_err == nil {
		return TagView(uni, tag_map)
	}
	tag_search_map, tag_search_err := routep.Comp("/tag-search/{slug}", uni.P)
	if tag_search_err == nil {
		return TagSearch(uni, tag_search_map)
	}
	content_map, content_err := routep.Comp("/{slug}", uni.P)
	if content_err == nil && len(content_map["slug"]) > 0 {
		return ContentView(uni, content_map)
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

func Index(uni *context.Uni) error {
	var v []interface{}
	var q m
	search_sl, has := uni.Req.Form["search"];
	if has && len(search_sl[0]) > 0 {
		q = m{"$and": content_model.GenerateQuery(search_sl[0])}
		uni.Dat["search"] = search_sl[0]
	}
	uni.Db.C("contents").Find(q).Sort("-created").All(&v)
	scut.Strify(v) // TODO: not sure this is needed now Inud handles `ObjectIdHex("blablabla")` ids well.
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/index"}
	return nil
}

func ListTags(uni *context.Uni) error {
	var v []interface{}
	uni.Db.C("tags").Find(nil).All(&v)
	scut.Strify(v) // TODO: not sure this is needed now Inud handles `ObjectIdHex("blablabla")` ids well.
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/tags"}
	return nil
}

func List(uni *context.Uni) error {
	ma, err := routep.Comp("/admin/content/list/{type}", uni.Req.URL.Path)
	if err != nil {
		return fmt.Errorf("Bad url at list.")
	}
	typ, has := ma["type"]
	if !has {
		return fmt.Errorf("Can not extract typ at list.")
	}
	var v []interface{}
	q := m{"type":typ}
	search_sl, has := uni.Req.Form["search"];
	if has && len(search_sl[0]) > 0 {
		q["$and"] = content_model.GenerateQuery(search_sl[0])
		uni.Dat["search"] = search_sl[0]
	}
	uni.Db.C("contents").Find(q).Sort("-created").All(&v)
	scut.Strify(v) // TODO: not sure this is needed now Inud handles `ObjectIdHex("blablabla")` ids well.
	uni.Dat["type"] = typ
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/list"}
	return nil
}

// Both everyone and personal.
func TypeConfig(uni *context.Uni) error {
	ma, err := routep.Comp("/admin/content/type-config/{type}", uni.Req.URL.Path)
	if err != nil {
		return fmt.Errorf("Bad url at type config.")
	}
	typ, has := ma["type"]
	if !has {
		return fmt.Errorf("Can not extract typ at type config.")
	}
	op, ok := jsonp.Get(uni.Opt, "Modules.content.types." + typ)
	if !ok {
		return fmt.Errorf("Can not find content type " + typ + " in options.")
	}
	uni.Dat["type"] = typ
	uni.Dat["type_options"], _ = json.MarshalIndent(op, "", "    ")
	uni.Dat["op"] = op
	user_type_op, has := jsonp.Get(uni.Dat["_user"], "content_options." + typ)
	uni.Dat["user_type_op"] = user_type_op
	uni.Dat["_points"] = []string{"content/type-config"}
	return nil
}

func Config(uni *context.Uni) error {
	op, _ := jsonp.Get(uni.Opt, "Modules.content")
	v, err := json.MarshalIndent(op, "", "    ")
	if err != nil {
		return fmt.Errorf("Can't marshal content options.")
	}
	uni.Dat["content_options"] = string(v)
	uni.Dat["_points"] = []string{"content/config"}
	return nil
}

// Called from both admin and outside editing.
func Edit(uni *context.Uni, ma map[string]string) error {
	typ, hast := ma["type"]
	if !hast {
		return fmt.Errorf("Can't extract type at edit.")
	}
	uni.Dat["content_type"] = typ
	rules, hasr := jsonp.GetM(uni.Opt, "Modules.content.types." + typ + ".rules")
	if !hasr {
		return fmt.Errorf("Can't find rules of " + typ)
	}
	uni.Dat["type"] = typ
	id, hasid := ma["id"]
	var indb interface{}
	if hasid {
		uni.Dat["op"] = "update"
		uni.Db.C("contents").Find(m{"_id": bson.ObjectIdHex(id)}).One(&indb)
		indb = basic.Convert(indb)
		resolver.ResolveOne(uni.Db, indb)
		uni.Dat["content"] = indb
	} else {
		uni.Dat["op"] = "insert"
	}
	f, ferr := scut.RulesToFields(rules, context.Convert(indb))
	if ferr != nil {
		return ferr
	}
	uni.Dat["fields"] = f
	return nil
}

// Admin edit
func AEdit(uni *context.Uni) error {
	ma, err := routep.Comp("/admin/content/edit/{type}/{id}", uni.Req.URL.Path)
	if err != nil {
		return fmt.Errorf("Bad url at edit.")
	}
	Edit(uni, ma)
	uni.Dat["_points"] = []string{"content/edit"}
	return nil
}

func AD(uni *context.Uni) error {
	var err error
	m, _ := routep.Comp("/admin/content/{view}", uni.Req.URL.Path)
	uni.Dat["content_menu"] = getSidebar(uni)
	switch m["view"] {
	case "":
		err = Index(uni)
	case "config":
		err = Config(uni)
	case "type-config":
		err = TypeConfig(uni)
	case "edit":
		err = AEdit(uni)
	case "list":
		err = List(uni)
	case "tags":
		err = ListTags(uni)
	default:
		err = fmt.Errorf("Unkown content view.")
	}
	return err
}
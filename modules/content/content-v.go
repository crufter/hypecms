package content

import(
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/routep"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/content/model"
	"github.com/opesun/hypecms/modules/display/model"
	"encoding/json"
	"fmt"
	"github.com/opesun/resolver"
	"strings"
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
	ed_err := Edit(uni, urimap)
	if ed_err != nil { return ed_err }
	uni.Dat["_hijacked"] = true
	uni.Dat["_points"] = []string{"edit-content"}	// Must contain require content/edit-form.t to work well.
	return nil
}

func TagView(uni *context.Uni, urimap map[string]string) error {
	fieldname := "slug"		// This should not be hardcoded.
	list, err := content_model.ListContentsByTag(uni.Db, fieldname, urimap["slug"])
	if err != nil {
		uni.Dat["error"] = err.Error()
	} else {
		display_model.CreateExcerpts(list, m{"content":float64(300)})
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
	visible_types := []string{}
	types, has := jsonp.GetM(uni.Opt, "Modules.content.types")
	if !has { return fmt.Errorf("Can't find content types.") }
	for i, _ := range types {
		visible_types = append(visible_types, i)
	}
	q := m{"type": m{"$in": visible_types}}
	search_sl, has := uni.Req.Form["search"];
	if has && len(search_sl[0]) > 0 {
		q["$and"] = content_model.GenerateQuery(search_sl[0])
		uni.Dat["search"] = search_sl[0]
	}
	skip_amount, paging := display_model.DoPaging(uni.Db, "contents", q, "page", map[string][]string(uni.Req.Form), uni.P + "?" + uni.Req.URL.RawQuery, 10)
	uni.Db.C("contents").Find(q).Sort("-created").Skip(skip_amount).Limit(10).All(&v)
	uni.Dat["paging"] = paging
	v = basic.Convert(v).([]interface{})
	content_model.ConnectWithDrafts(uni.Db, v)
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
	skip_amount, paging := display_model.DoPaging(uni.Db, "contents", q, "page", map[string][]string(uni.Req.Form), uni.P + "?" + uni.Req.URL.RawQuery, 10)
	uni.Db.C("contents").Find(q).Sort("-created").Skip(skip_amount).Limit(10).All(&v)
	uni.Dat["paging"] = paging
	v = basic.Convert(v).([]interface{})
	content_model.ConnectWithDrafts(uni.Db, v)
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

func EditContent(uni *context.Uni, typ, id string, hasid bool) (interface{}, error) {
	uni.Dat["is_content"] = true
	var indb interface{}
	if hasid {
		uni.Dat["op"] = "update"
		uni.Db.C("contents").Find(m{"_id": bson.ObjectIdHex(id)}).One(&indb)						// Ugly.
		indb = basic.Convert(indb)
		resolver.ResolveOne(uni.Db, indb, nil)
		uni.Dat["content"] = indb
		latest_draft := content_model.GetUpToDateDraft(uni.Db, bson.ObjectIdHex(id), indb.(map[string]interface{}))
		scut.Strify(latest_draft)
		uni.Dat["latest_draft"] = latest_draft
	} else {
		uni.Dat["op"] = "insert"
	}
	return context.Convert(indb), nil
}

func EditDraft(uni *context.Uni, typ, id string, hasid bool) (interface{}, error) {
	uni.Dat["is_draft"] = true
	if hasid {
		built, err := content_model.BuildDraft(uni.Db, typ + "_draft", id)
		if err != nil { return nil, err }
		d := built["data"].(map[string]interface{})
		if content_model.HasContentParent(built) {
			uni.Dat["content_parent"] = true
			uni.Dat["up_to_date"] = content_model.IsDraftUpToDate(uni.Db, built, d)
			uni.Dat["op"] = "update"
		} else {	// It's possible that it has no parent at all, then it is a fresh new draft, first version.
			uni.Dat["op"] = "insert"
		}
		resolver.ResolveOne(uni.Db, d, nil)
		scut.Strify(d)
		scut.Strify(built)
		uni.Dat["content"] = d
		uni.Dat["draft"] = built
		return d, nil
	}
	uni.Dat["op"] = "insert"
	return map[string]interface{}{}, nil
}

// You don't actually edit anything on a past version...
func EditVersion(uni *context.Uni, typ, id string) (interface{}, error) {
	return nil, nil
}

// Ex: realType of "blog_draft" is "blog".
func realType(typ string) string {
	li := strings.LastIndex(typ, "_")
	if li != -1 {
		return typ[0:li]
	}
	return typ
}

// Ex: subType of "blog_draft" is "draft", subtype of "blog" is "content".
func subType(typ string) string {
	li := strings.LastIndex(typ, "_")
	if li != -1 {
		return typ[li+1:]
	}
	return "content"
}

// Called from both admin and outside editing.
// ma containts type and id members extracted out of the url.
func Edit(uni *context.Uni, ma map[string]string) error {
	typ, hast := ma["type"]
	rtyp := realType(typ)
	if !hast {
		return fmt.Errorf("Can't extract type at edit.")
	}
	rules, hasr := jsonp.GetM(uni.Opt, "Modules.content.types." + rtyp + ".rules")
	if !hasr {
		return fmt.Errorf("Can't find rules of " + rtyp)
	}
	uni.Dat["content_type"] = rtyp
	uni.Dat["type"] = rtyp
	id, hasid := ma["id"]
	var field_dat interface{}
	var err error
	subt := subType(typ)
	switch subt {
	case "content":
		field_dat, err = EditContent(uni, typ, id, hasid)
	case "draft":
		field_dat, err = EditDraft(uni, rtyp, id, hasid)
	case "version":
		if !hasid { return fmt.Errorf("Version must have id.") }
		field_dat, err = EditVersion(uni, rtyp, id)
	default:
		panic(fmt.Sprintf("Unkown content subtype: %v.", subt))
	}
	if err != nil { return err }
	fields, err := scut.RulesToFields(rules, field_dat)
	if err != nil { return err }
	uni.Dat["fields"] = fields
	return nil
}

// Admin edit
func AEdit(uni *context.Uni) error {
	ma, err := routep.Comp("/admin/content/edit/{type}/{id}", uni.Req.URL.Path)
	if err != nil {
		return fmt.Errorf("Bad url at edit.")
	}
	ed_err := Edit(uni, ma)
	if ed_err != nil { return ed_err }
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
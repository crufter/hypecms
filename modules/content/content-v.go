package content

import (
	"encoding/json"
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/hypecms/model/patterns"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/content/model"
	"github.com/opesun/hypecms/modules/display/model"
	"github.com/opesun/jsonp"
	"github.com/opesun/resolver"
	"github.com/opesun/routep"
	"labix.org/v2/mgo/bson"
	"strings"
)

type m map[string]interface{}

func userEdit(uni *context.Uni, urimap map[string]string) error {
	ed_err := Edit(uni, urimap)
	if ed_err != nil {
		return ed_err
	}
	uni.Dat["_points"] = []string{"edit-content"} // Must contain require content/edit-form.t to work well.
	return nil
}

func tagView(uni *context.Uni, urimap map[string]string) error {
	fieldname := "slug" // This should not be hardcoded.
	specific := len(urimap) == 2
	var search_value string
	var children_query map[string]interface{}
	if specific {
		children_query["type"] = urimap["first"]
		search_value = urimap["second"]
	} else {
		search_value = urimap["first"]
	}
	tag, err := content_model.FindTag(uni.Db, fieldname, search_value)
	if err != nil {
		return err
	}
	pnq := uni.P + "?" + uni.Req.URL.RawQuery
	query := map[string]interface{}{
		"ex": map[string]interface{}{
			"content": 300,
		},
		"so": "-created",
		"c":  "contents",
		"q": map[string]interface{}{
			"_tags": tag["_id"],
		},
		"p": "page",
		"l": 20,
	}
	cl := display_model.RunQuery(uni.Db, "content_list", query, uni.Req.Form, pnq)
	uni.Dat["content_list"] = cl["content_list"]
	uni.Dat["content_list_navi"] = cl["content_list_navi"]
	uni.Dat["_points"] = []string{"tag"}
	return nil
}

func tagSearch(uni *context.Uni) error {
	var name_search string
	search, has := uni.Req.Form["search"]
	if has {
		name_search = search[0]
	}
	q := content_model.TagSearchQuery("name", name_search)
	pnq := uni.P + "?" + uni.Req.URL.RawQuery
	query := map[string]interface{}{
		"so": "-created",
		"c":  "tags",
		"q":  q,
		"p":  "page",
		"l":  20,
	}
	cl := display_model.RunQuery(uni.Db, "tag_list", query, uni.Req.Form, pnq)
	uni.Dat["tag_list"] = cl["tag_list"]
	uni.Dat["search_term"] = name_search
	uni.Dat["tag_list_navi"] = cl["tag_list_navi"]
	uni.Dat["_points"] = []string{"tag-search"}
	return nil
}

func contentView(uni *context.Uni, content_map map[string]string) error {
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
	if !found {
		return fmt.Errorf("Can't find content.")
	}
	dont_query := map[string]interface{}{"password": 0}
	resolver.ResolveOne(uni.Db, content, dont_query)
	uni.Dat["_points"] = []string{"content"}
	uni.Dat["content"] = content
	return nil
}

func contentSearch(uni *context.Uni) error {
	q := map[string]interface{}{}
	search_sl, has := uni.Req.Form["search"]
	var search_term string
	if has && len(search_sl[0]) > 0 {
		search_term = search_sl[0]
		q["$and"] = content_model.GenerateQuery(search_term)
		uni.Dat["search"] = search_sl[0]
	}
	query := map[string]interface{}{
		"ex": map[string]interface{}{
		"content": 300,
		},
		"so": "-created",
		"c":  "contents",
		"q":  q,
		"p":  "page",
		"l":  20,
	}
	pnq := uni.P + "?" + uni.Req.URL.RawQuery
	cl := display_model.RunQuery(uni.Db, "content_list", query, uni.Req.Form, pnq)
	fmt.Println(cl)
	uni.Dat["content_list"] = cl["content_list"]
	uni.Dat["search_term"] = search_term
	uni.Dat["content_list_navi"] = cl["content_list_navi"]
	uni.Dat["_points"] = []string{"content-search"}
	return nil
}

func Front(uni *context.Uni, hijacked *bool) error {
	edit_map, edit_err := routep.Comp("/content/edit/{type}/{id}", uni.P)
	if edit_err == nil {
		*hijacked = true
		return userEdit(uni, edit_map)
	}
	tag_map, tag_err := routep.Comp("/tag/{first}/{second}", uni.P)
	// Tag view: list contents in that category.
	if tag_err == nil {
		*hijacked = true
		return tagView(uni, tag_map)
	}
	_, tag_search_err := routep.Comp("/tag-search", uni.P)
	if tag_search_err == nil {
		*hijacked = true
		return tagSearch(uni)
	}
	_, content_search_err := routep.Comp("/content-search", uni.P)
	if content_search_err == nil {
		*hijacked = true
		return contentSearch(uni)
	}
	content_map, content_err := routep.Comp("/{slug}", uni.P)
	if content_err == nil && len(content_map["slug"]) > 0 {
		*hijacked = true
		return contentView(uni, content_map)
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
	if !has {
		return fmt.Errorf("Can't find content types.")
	}
	for i, _ := range types {
		visible_types = append(visible_types, i)
	}
	q := m{"type": m{"$in": visible_types}}
	search_sl, has := uni.Req.Form["search"]
	if has && len(search_sl[0]) > 0 {
		q["$and"] = content_model.GenerateQuery(search_sl[0])
		uni.Dat["search"] = search_sl[0]
	}
	paging_inf := display_model.DoPaging(uni.Db, "contents", q, "page", map[string][]string(uni.Req.Form), uni.P+"?"+uni.Req.URL.RawQuery, 10)
	uni.Db.C("contents").Find(q).Sort("-created").Skip(paging_inf.Skip).Limit(10).All(&v)
	uni.Dat["paging"] = paging_inf
	v = basic.Convert(v).([]interface{})
	content_model.HaveUpToDateDrafts(uni.Db, v)
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/index"}
	return nil
}

// This functionality is almost the same as the tagSearch on the outside :S
func ListTags(uni *context.Uni) error {
	var v []interface{}
	uni.Db.C("tags").Find(nil).All(&v)
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/tags"}
	return nil
}

// Almost the same as contentSearc as outside :S
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
	q := m{"type": typ}
	search_sl, has := uni.Req.Form["search"]
	if has && len(search_sl[0]) > 0 {
		q["$and"] = content_model.GenerateQuery(search_sl[0])
		uni.Dat["search"] = search_sl[0]
	}
	paging_inf := display_model.DoPaging(uni.Db, "contents", q, "page", map[string][]string(uni.Req.Form), uni.P+"?"+uni.Req.URL.RawQuery, 10)
	uni.Db.C("contents").Find(q).Sort("-created").Skip(paging_inf.Skip).Limit(10).All(&v)
	uni.Dat["paging"] = paging_inf
	v = basic.Convert(v).([]interface{})
	content_model.HaveUpToDateDrafts(uni.Db, v)
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
	op, ok := jsonp.Get(uni.Opt, "Modules.content.types."+typ)
	if !ok {
		return fmt.Errorf("Can not find content type " + typ + " in options.")
	}
	uni.Dat["type"] = typ
	uni.Dat["type_options"], _ = json.MarshalIndent(op, "", "    ")
	uni.Dat["op"] = op
	user_type_op, has := jsonp.Get(uni.Dat["_user"], "content_options."+typ)
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
		uni.Db.C("contents").Find(m{"_id": bson.ObjectIdHex(id)}).One(&indb) // Ugly.
		indb = basic.Convert(indb)
		resolver.ResolveOne(uni.Db, indb, nil)
		uni.Dat["content"] = indb
		latest_draft := content_model.GetUpToDateDraft(uni.Db, bson.ObjectIdHex(id), indb.(map[string]interface{}))
		uni.Dat["latest_draft"] = latest_draft
		timeline, err := content_model.ContentTimeline(uni.Db, indb.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		uni.Dat["timeline"] = timeline
	} else {
		uni.Dat["op"] = "insert"
	}
	return context.Convert(indb), nil
}

func EditDraft(uni *context.Uni, typ, id string, hasid bool) (interface{}, error) {
	uni.Dat["is_draft"] = true
	if hasid {
		built, err := content_model.BuildDraft(uni.Db, typ, id)
		if err != nil {
			return nil, err
		}
		d := built["data"].(map[string]interface{})
		if _, draft_of_sg := built["draft_of"]; draft_of_sg {
			uni.Dat["content_parent"] = true
			fresher, err := content_model.IsDraftUpToDate(uni.Db, built, d)
			if err != nil {
				return nil, err
			}
			uni.Dat["up_to_date"] = fresher
			uni.Dat["op"] = "update"
		} else { // It's possible that it has no parent at all, then it is a fresh new draft, first version.
			uni.Dat["op"] = "insert"
		}
		resolver.ResolveOne(uni.Db, d, nil)
		uni.Dat["content"] = d
		timeline, err := content_model.DraftTimeline(uni.Db, patterns.ToIdWithCare(id))
		if err != nil {
			return nil, err
		}
		uni.Dat["timeline"] = timeline
		uni.Dat["draft"] = built
		return d, nil
	}
	uni.Dat["op"] = "insert"
	return map[string]interface{}{}, nil
}

// You don't actually edit anything on a past version...
func EditVersion(uni *context.Uni, typ, id string) (interface{}, error) {
	uni.Dat["is_version"] = true
	version_id := patterns.ToIdWithCare(id)
	version, err := content_model.FindVersion(uni.Db, version_id)
	if err != nil {
		return nil, err
	}
	resolver.ResolveOne(uni.Db, version, nil)
	timeline, err := content_model.DraftTimeline(uni.Db, version_id)
	if err != nil {
		return nil, err
	}
	uni.Dat["timeline"] = timeline
	uni.Dat["op"] = "update"
	uni.Dat["content"] = version
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
	rules, hasr := jsonp.GetM(uni.Opt, "Modules.content.types."+rtyp+".rules")
	if !hasr {
		return fmt.Errorf("Can't find rules of " + rtyp)
	}
	uni.Dat["content_type"] = rtyp
	uni.Dat["type"] = rtyp
	id, ok := ma["id"]
	hasid := ok && len(id) > 0 // Corrigate routep.Comp because it sets a map key with an empty value...
	var field_dat interface{}
	var err error
	subt := subType(typ)
	switch subt {
	case "content":
		field_dat, err = EditContent(uni, typ, id, hasid)
	case "draft":
		fmt.Println(rtyp, id, hasid)
		field_dat, err = EditDraft(uni, rtyp, id, hasid)
	case "version":
		if !hasid {
			return fmt.Errorf("Version must have id.")
		}
		field_dat, err = EditVersion(uni, rtyp, id)
	default:
		panic(fmt.Sprintf("Unkown content subtype: %v.", subt))
	}
	if err != nil {
		return err
	}
	fields, err := scut.RulesToFields(rules, field_dat)
	if err != nil {
		return err
	}
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
	if ed_err != nil {
		return ed_err
	}
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

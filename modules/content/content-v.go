package content

import(
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/routep"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/hypecms/api/scut"
	"github.com/opesun/hypecms/modules/content/model"
	"encoding/json"
	"fmt"
)

func Front(uni *context.Uni) error {
	ed, ed_err := routep.Comp("/admin/content/edit/{type}/{id}", uni.P)
	if ed_err == nil {
		ulev, hasu := jsonp.GetI(uni.Opt, "_user.level")
		if !hasu {
			return fmt.Errorf("No user level found, or it is not an integer.")
		}
		_, hasid := ed["id"]
		if hasid && ulev < minLev(uni.Opt, "edit") {
			return fmt.Errorf("You have no rights to edit a content.")
		} else if ulev < minLev(uni.Opt, "insert") {
			return fmt.Errorf("You have no rights to insert a content.")
		}
		uni.Dat["_hijacked"] = true
		Edit(uni, ed)
		uni.Dat["_points"] = []string{"edit-content"}	// Must contain require content/edit-form.t to work well.
		return nil
	}
	m, err := routep.Comp("/{slug}", uni.P)
	if err == nil && len(m["slug"]) > 0 {
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
		content, found := content_model.FindContent(uni.Db, slug_keys, m["slug"])
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

func Index(uni *context.Uni) error {
	var v []interface{}
	uni.Db.C("contents").Find(nil).Sort("-created").All(&v)
	scut.Strify(v) // TODO: not sure this is needed now Inud handles `ObjectIdHex("blablabla")` ids well.
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/index"}
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
	uni.Db.C("contents").Find(m{"type":typ}).Sort("-created").All(&v)
	scut.Strify(v) // TODO: not sure this is needed now Inud handles `ObjectIdHex("blablabla")` ids well.
	uni.Dat["latest"] = v
	uni.Dat["_points"] = []string{"content/list"}
	return nil
}

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
	rules, hasr := jsonp.Get(uni.Opt, "Modules.content.types." + typ + ".rules")
	if !hasr {
		return fmt.Errorf("Can't find rules of " + typ)
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
	f, ferr := scut.RulesToFields(rules, context.Convert(indb))
	if ferr != nil {
		return ferr
	}
	uni.Dat["fields"] = f
	return nil
}

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
	default:
		err = fmt.Errorf("Unkown content view.")
	}
	return err
}
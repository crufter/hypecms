// Package scut contains a somewhat ugly but useful collection of frequently appearing patterns to allow faster prototyping.
package scut

import(
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"launchpad.net/mgo/bson"
	"launchpad.net/mgo"
	"time"
	"github.com/opesun/jsonp"
	"sort"
	"path/filepath"
)

// Insert, and update/delete, by Id.
// TODO: uni should definitely not be passed here. I think we should create a communication channel of some sort to allow these inner functions to
// notify the outside world about the happenings. This could be done with some kind of event emitter/receiver, or simply with callbacks... etc,
// but this "even the deepest parts of the architecture has access to everything" is just plain wrong.
func Inud(uni *context.Uni, dat map[string]interface{}, coll, op, id string) error {
	db := uni.Db
	var err error
	if (op == "update" || op == "delete") && len(id) != 24 {
		if len(id) == 39 {
			id = id[13:37]
		} else {
			return fmt.Errorf("Length of id is not 24 or 39 at updating or deleting.")
		}
	}
	switch op {
	case "insert":
		err = db.C(coll).Insert(dat)
	case "update":
		err = db.C(coll).Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": dat})
	case "delete":
		var v interface{}
		err = db.C(coll).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&v)
		if v == nil {
			return fmt.Errorf("Can't find document " + id + " in " + coll)
		}
		if err != nil {
			return err
		}
		// Transactions would not hurt here, but maybe we can get away with upserts.
		_, err = db.C(coll + "_deleted").Upsert(bson.M{"_id": bson.ObjectIdHex(id)}, v)
		if err != nil {
			return err
		}
		err = db.C(coll).Remove(bson.M{"_id": bson.ObjectIdHex(id)})
		if err != nil {
			return err
		}
	case "restore":
		// err = db.C(coll).Find
		// Not implemented yet.
	}
	if err != nil {
		return err
	} else {
		hooks, has := jsonp.Get(uni.Opt, "Hooks." + coll + "." + op)
		if has {
			for _, v := range hooks.([]string) {
				//f := mod.GetHook(
				fmt.Println(v)
			}
		}
	}
	return nil
}

// Iterates a [] coming from a mgo query and converts the "_id" members from bson.ObjectId to string.
// TODO: not sure this is needed now Inud handles `ObjectIdHex("blablabla")` ids well.
func Strify(v []interface{}) {
	for _, val := range v {
		val.(bson.M)["_id"] = val.(bson.M)["_id"].(bson.ObjectId).Hex()
	}
}

// Called before install and uninstall automatically, and you must call it by hand every time you want to modify the option document.
func CreateOptCopy(db *mgo.Database) bson.ObjectId {
	var v interface{}
	db.C("options").Find(nil).Sort(bson.M{"created": -1}).Limit(1).One(&v)
	ma := v.(bson.M)
	ma["_id"] = bson.NewObjectId()
	ma["created"] = time.Now().Unix()
	db.C("options").Insert(ma)
	return ma["_id"].(bson.ObjectId)
}

// A more generic version of abcKeys. Takes a map[string]interface{} and puts every element of that into an []interface{}, ordered by keys alphabetically.
// TODO: find the intersecting parts between the two functions and refactor.
func OrderKeys(d map[string]interface{}) []interface{} {
	keys := []string{}
	for i, _ := range d {
		keys = append(keys, i)
	}
	sort.Strings(keys)
	ret := []interface{}{}
	for _, v := range keys {
		if ma, is_ma := d[v].(map[string]interface{}); is_ma {
			// RETHINK: What if a key field gets overwritten? Should we name it _key?
			ma["key"] = v
		}
		ret = append(ret, d[v])
	}
	return ret
}

// Takes a dat map[string]interface{}, and puts every element of that which is defined in r to a slice, sorted by the keys ABC order.
// prior parameter can override the default abc ordering, so keys in prior will be the first ones in the slice, if those keys exist.
func abcKeys(rule map[string]interface{}, dat map[string]interface{}, prior []string) []map[string]interface{} {
	ret := []map[string]interface{}{}
	already_in := map[string]struct{}{}
	for _, v := range prior {
		if _, contains := rule[v]; contains {
			item := map[string]interface{}{"key":v}
			if dat != nil {
				item["value"] = dat[v]
			}
			ret = append(ret, item)
			already_in[v] = struct{}{}
		}
	}
	keys := []string{}
	for i, _ := range rule {
		keys = append(keys, i)
	}
	sort.Strings(keys)
	for _, v := range keys {
		if _, in := already_in[v]; !in {
			item := map[string]interface{}{"key":v}
			if dat != nil {
				item["value"] = dat[v]
			}
			ret = append(ret, item)
		}
	}
	return ret
}

// Takes an extraction/validation rule, a document and from that creates a slice which can be easily displayed by a templating engine as a html form.
func RulesToFields(rule interface{}, dat interface{}) ([]map[string]interface{}, error) {
	rm, rm_ok := rule.(map[string]interface{})
	if !rm_ok {
		return nil, fmt.Errorf("Rule is not a map[string]interface{}.")
	}
	datm, datm_ok := dat.(map[string]interface{})
	if !datm_ok && dat != nil {
		return nil, fmt.Errorf("Dat is not a map[string]interface{}.")
	}
	return abcKeys(rm, datm, []string{"title", "name", "slug"}), nil
}

func TemplateType(opt map[string]interface{}) string {
	_, priv := opt["TplIsPrivate"]
	var ttype string
	if priv {
		ttype = "private"
	} else {
		ttype = "public"
	}
	return ttype
}

func TemplateName(opt map[string]interface{}) string {
	tpl, has_tpl := opt["Template"]
	if !has_tpl {
		tpl = "default"
	}
	return tpl.(string)
}

// Observes opt and gives you back a string describing the path of your template eg "templates/public/template_name"
func GetTPath(opt map[string]interface{}) string {
	templ := TemplateName(opt)
	ttype := TemplateType(opt)
	return filepath.Join("templates", ttype, templ)
}

// Calculate missing fields, we compare dat to r.
func CalcMiss(rule map[string]interface{}, dat map[string]interface{}) []string {
	missing_fields := []string{}
	for i, _ := range rule {
		if _, ex := dat[i]; !ex {
			missing_fields = append(missing_fields, i)
		}
	}
	return missing_fields
}
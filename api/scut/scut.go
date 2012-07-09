// Package scut contains a somewhat ugly but useful collection of frequently appearing patterns to allow faster prototyping.
package scut

import(
	"fmt"
	"labix.org/v2/mgo/bson"
	"sort"
	"path/filepath"
)

// Iterates a [] coming from a mgo query and converts the "_id" members from bson.ObjectId to string.
// TODO: not sure this is needed now Inud handles `ObjectIdHex("blablabla")` ids well.
func Strify(v []interface{}) {
	for _, val := range v {
		val.(bson.M)["_id"] = val.(bson.M)["_id"].(bson.ObjectId).Hex()
	}
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
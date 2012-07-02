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
)

// Insert, and update/delete, by Id.
func Inud(uni *context.Uni, dat map[string]interface{}, coll, op, id string) error {
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
		err = uni.Db.C(coll).Insert(dat)
	case "update":
		err = uni.Db.C(coll).Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": dat})
	case "delete":
		var v interface{}
		err = uni.Db.C(coll).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&v)
		if v == nil {
			return fmt.Errorf("Can't find document " + id + " in " + coll)
		}
		if err != nil {
			return err
		}
		// Transactions would not hurt here, but maybe we can get away with upserts.
		_, err = uni.Db.C(coll + "_deleted").Upsert(bson.M{"_id": bson.ObjectIdHex(id)}, v)
		if err != nil {
			return err
		}
		err = uni.Db.C(coll).Remove(bson.M{"_id": bson.ObjectIdHex(id)})
		if err != nil {
			return err
		}
	case "restore":
		// err = uni.Db.C(coll).Find
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

// Takes a map[string]interface{}, and puts every element of that to a slice, sorted by the keys ABC order.
// prior parameter can override the default abc ordering, so keys in prior will be the first ones in the slice, if those keys exist.
func abcKeys(r map[string]interface{}, dat map[string]interface{}, prior []string) []map[string]interface{} {
	ret := []map[string]interface{}{}
	already_in := map[string]struct{}{}
	for _, v := range prior {
		if _, contains := r[v]; contains {
			val := map[string]interface{}{"key":v, "value": dat[v]}
			ret = append(ret, val)
			already_in[v] = struct{}{}
		}
	}
	keys := []string{}
	for i, _ := range r {
		keys = append(keys, i)
	}
	sort.Strings(keys)
	for _, v := range keys {
		if _, in := already_in[v]; !in {
			val := map[string]interface{}{"key":v, "value": dat[v]}
			ret = append(ret, val)
		}
	}
	return ret
}

// Takes an extraction/validation rule, a document and from that creates a slice which can be easily displayed by a templating engine as a html form.
func RulesToFields(r interface{}, dat interface{}) ([]map[string]interface{}, error) {
	rm, rm_ok := r.(map[string]interface{})
	if !rm_ok {
		return nil, fmt.Errorf("Rule is not a map[string]interface{}.")
	}
	datm, datm_ok := dat.(map[string]interface{})
	if !datm_ok {
		return nil, fmt.Errorf("Dat is not a map[string]interface{}.")
	}
	return abcKeys(rm, datm, []string{"title", "name", "slug"}), nil
}
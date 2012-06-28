// Package scut contains a somewhat ugly but useful collection of frequently appearing patterns to allow faster prototyping.
package scut

import(
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"launchpad.net/mgo/bson"
	"launchpad.net/mgo"
	"time"
	"github.com/opesun/jsonp"
)

// Insert, update, delete.
func Inud(uni *context.Uni, dat map[string]interface{}, res *map[string]interface{}, coll, op, id string) error {
	var err error
	if (op == "update" || op == "delete") && len(id) != 24 {
		if len(id) == 39 {
			id = id[13:37]
		} else {
			return fmt.Errorf("Length of id is 0 at updating or deleting.")
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
		(*res)["success"] = false
		(*res)["reason"] = err.Error()
	} else {
		(*res)["success"] = true
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
func Strify(v []interface{}) {
	for _, val := range v {
		val.(bson.M)["_id"] = val.(bson.M)["_id"].(bson.ObjectId).Hex()
	}
}

func CreateOptCopy(db *mgo.Database) bson.ObjectId {
	var v interface{}
	db.C("options").Find(nil).Sort(bson.M{"created": -1}).Limit(1).One(&v)
	ma := v.(bson.M)
	ma["_id"] = bson.NewObjectId()
	ma["created"] = time.Now().Unix()
	db.C("options").Insert(ma)
	return ma["_id"].(bson.ObjectId)
}
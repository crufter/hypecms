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
	if (op == "update" || op == "delete") || len(id) == 0 {
		return fmt.Errorf("Length of id is 0 az updating/deleting.")
	}
	switch op {
	case "insert":
		err = uni.Db.C(coll).Insert(dat)
	case "update":
		err = uni.Db.C(coll).Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": dat})
	case "delete":
		err = uni.Db.C(coll).Remove(bson.M{"_id": bson.ObjectIdHex(id)})
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
	return err
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
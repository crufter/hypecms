package basic

import(
	ifaces "github.com/opesun/hypecms/interfaces"
	"launchpad.net/mgo/bson"
	"launchpad.net/mgo"
	"fmt"
)

// by Id.
func Find(db *mgo.Database, coll, id string) map[string]interface{} {
	var v interface{}
	db.C("users").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&v)
	if v != nil {
		return Convert(v).(map[string]interface{})
	}
	return nil
}

// Insert, and update/delete, by Id.
func Inud(db *mgo.Database, ev ifaces.Event, dat map[string]interface{}, coll, op, id string) error {
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
	ev.Trigger(coll + "." + op, dat)
	return nil
}

// Convert multiply nested bson.M-s to map[string]interface{}
// Written by Rog Peppe.
func Convert(x interface{}) interface{} {
	if x, ok := x.(bson.M); ok {
		for key, val := range x {
			x[key] = Convert(val)
		}
		return (map[string]interface{})(x)
	}
	return x
}
package basic

import(
	ifaces "github.com/opesun/hypecms/interfaces"
	"labix.org/v2/mgo/bson"
	"labix.org/v2/mgo"
	"fmt"
	"time"
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

func CreateCopy(db *mgo.Database, collname, sortfield string) bson.ObjectId {
	var v interface{}
	db.C(collname).Find(nil).Sort(sortfield).Limit(1).One(&v)
	ma := v.(bson.M)
	ma["_id"] = bson.NewObjectId()
	ma["created"] = time.Now().Unix()
	db.C(collname).Insert(ma)
	return ma["_id"].(bson.ObjectId)
}

// Called before install and uninstall automatically, and you must call it by hand every time you want to modify the option document.
func CreateOptCopy(db *mgo.Database) bson.ObjectId {
	return CreateCopy(db, "options", "-created")
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
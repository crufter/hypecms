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

func Convert(x interface{}) interface{} {
	if y, ok := x.(bson.M); ok {
		for key, val := range y {
			y[key] = Convert(val)
		}
		return (map[string]interface{})(y)
	} else if z, ok := x.([]interface{}); ok {
		for i, v := range z {
			z[i] = Convert(v)
		}
	}
	return x
}

func convertMapId(x map[string]interface{}) {
	if id, has_id := x["_id"]; has_id {
		x["_id"] = id.(bson.ObjectId).Hex()
	}
}

// Used in views, where we dont want to show display ObjectIdHex("ab889d8ec889") but rather "ab889d8ec889".
// x must be a result map[string]interface{} or result []interface{}
func IdsToStrings(x interface{}) {
	if ma, ok := x.(map[string]interface{}); ok {
		convertMapId(ma)
	} else if ma_sl, sl_ok := x.([]interface{}); sl_ok {
		for _, v := range ma_sl {
			convertMapId(v.(map[string]interface{}))
		}
	} else {
		panic("Wrong input for IdsToString.")
	}
}

// Takes a query map and converts the "_id" member of it from string to bson.ObjectId
func StringToId(query map[string]interface{}) {
	if id_i, has := query["_id"]; has {
		if id_s, is_str := id_i.(string); is_str {
			query["_id"] = bson.ObjectIdHex(id_s)
		}
	}
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

// Kind of an extension of the extract module.
// Handles author fields 	(created_by, last_modified_by),
// And time fields			(created, last_modified)
func DateAndAuthor(rule map[string]interface{}, dat map[string]interface{}, user_id bson.ObjectId) {
	var inserting, updating bool
	if _, has_id := dat["id"]; has_id {
		updating = true
	} else {
		inserting = true
	}
	for i, _ := range rule {
		switch i {
		case "created_by":
			if inserting {
				dat[i] = user_id
			}
		case "last_modified_by":
			if updating {
				dat[i] = user_id
			}
		case "created":
			if inserting {
				dat[i] = time.Now()
			}
		case "last_modified":
			if updating {
				dat[i] = time.Now()
			}
		}
	}
}

func ExtractIds(dat map[string][]string, keys []string) ([]string, error) {
	ret := []string {}
	for _, v := range keys {
		id_s, has_id_s := dat[v]
		if !has_id_s {
			return nil, fmt.Errorf(v, " id is not found.")
		}
		id := id_s[0]
		if len(id) == 24 {
			ret = append(ret, v)
		} else if len(id) == 39 {
			ret = append(ret, id[13:37])
		} else {
			return nil, fmt.Errorf(v, " is not a proper id.")
		}
	}
	return ret, nil
}
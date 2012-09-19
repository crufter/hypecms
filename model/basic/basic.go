// A collection of independent helper functions mainly related to database operations.
package basic

import (
	"fmt"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/slugify"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

const (
	Delete_collection_postfix  = "_deleted"
	Version_collection_postfix = "_version"
	Created_by                 = "_users_created_by"
	Created                    = "created"
	Last_modified_by           = "_users_last_modified_by"
	Last_modified              = "last_modified"
	Version_datefield          = "version_date"
	Fresh                      = "fresh"            // Saved into a version, pointing to the "living" doc.
	Prev_version               = "previous_version" // Goes into the "living" doc.	
)

// Converts a given interface value to an ObjectId with utmost care, taking all possible malformedness into account.
func ToIdWithCare(id interface{}) bson.ObjectId {
	switch val := id.(type) {
	case bson.ObjectId:
	case string:
		id = bson.ObjectIdHex(StripId(val))
	default:
		panic(fmt.Sprintf("Can't create bson.ObjectId out of %T", val))
	}
	return id.(bson.ObjectId)
}

// Find by Id.
func Find(db *mgo.Database, coll string, id interface{}) map[string]interface{} {
	bson_id := ToIdWithCare(id)
	var v interface{}
	q := bson.M{"_id": bson_id}
	db.C(coll).Find(q).One(&v)
	if v != nil {
		return Convert(v).(map[string]interface{})
	}
	return nil
}

// Insert, and update/delete, by Id.
func Inud(db *mgo.Database, ev ifaces.Event, dat map[string]interface{}, coll, op, id string) error {
	return InudOpt(db, ev, dat, coll, op, id, false)
}

// Same as Inud, but takes into account versioning and drafts.
func InudVersion(db *mgo.Database, ev ifaces.Event, dat map[string]interface{}, coll, op, id string) error {
	return InudOpt(db, ev, dat, coll, op, id, true)
}

// TODO:
// with_upsert: true at deletion (you dont care if it is already copied to the deleted collection,
// false at restoration: you dont want to overwrite any document in the actual live collection.
func Copy(db *mgo.Database, from_collname, to_collname string, id bson.ObjectId) error {
	var v interface{}
	err := db.C(from_collname).Find(bson.M{"_id": id}).One(&v)
	if v == nil {
		return fmt.Errorf("Can't find document with id %v in collection %v.", id, from_collname)
	}
	if err != nil {
		return err
	}
	// Transactions would not hurt here, but maybe we can get away with upserts.
	q := bson.M{"_id": id}
	_, err = db.C(to_collname).Upsert(q, v)
	return err
}

// Moves document from one collection to another.
func Move(db *mgo.Database, from_collname, to_collname string, id bson.ObjectId) error {
	err := Copy(db, from_collname, to_collname, id)
	if err != nil {
		return err
	}
	q := bson.M{"_id": id}
	return db.C(from_collname).Remove(q)
}

// Deletes a document from a given collection by moving it to collname + "_deleted".
func Delete(db *mgo.Database, collname string, id bson.ObjectId) error {
	return Move(db, collname, collname+Delete_collection_postfix, id)
}

// Version id for insert, live id for querying.
func SaveVersion(db *mgo.Database, coll string, version_id, live_id, parent, root bson.ObjectId) error {
	var v interface{}
	err := db.C(coll).Find(bson.M{"_id": live_id}).One(&v)
	if err != nil {
		return err
	}
	copy := v.(bson.M)
	if parent != "" {
		copy["-parent"] = parent // This can be either a version, or a draft.
	}
	if root != "" {
		copy["root"] = root // This can be either a version, or a draft.
	}
	copy[Version_datefield] = time.Now().Unix()
	copy["_id"] = version_id
	return db.C(coll + Version_collection_postfix).Insert(copy)
}

// Query draft by id. Parent will be itself the draft, but we must extract the root from the draft.
// If it has none, he will become the root.
// Return the last_version of the parent draft too (used at content/model/versions.go)
func GetDraftParent(db *mgo.Database, coll string, draft_id bson.ObjectId) (parent, root, last_version bson.ObjectId, err error) {
	parent = draft_id
	var v interface{}
	err = db.C(coll + "_draft").Find(bson.M{"_id": draft_id}).One(&v)
	if err != nil {
		return
	}
	draft := v.(bson.M)
	// Not all drafts have the draft_of_version field.
	last_version_i, has_v := draft["draft_of_version"]
	if has_v {
		last_version = last_version_i.(bson.ObjectId)
	}
	if roo, has_roo := draft["root"]; has_roo {
		root = roo.(bson.ObjectId)
	} else {
		root = parent
	}
	return
}

// Query content by id. Get "pointing_to" field, query that version.
// Return the version as parent.
// Return its root as root, or itself as root if it has none.
func GetParentTroughContent(db *mgo.Database, coll string, content_id bson.ObjectId) (parent, root bson.ObjectId, err error) {
	var content_doc_i interface{}
	err = db.C(coll).Find(bson.M{"_id": content_id}).One(&content_doc_i)
	if err != nil {
		return
	}
	content_doc := content_doc_i.(bson.M)
	var v interface{}
	err = db.C(coll + Version_collection_postfix).Find(bson.M{"_id": content_doc["pointing_to"].(bson.ObjectId)}).One(&v)
	if err != nil {
		return
	}
	parent_version := v.(bson.M)
	parent = parent_version["_id"].(bson.ObjectId)
	roo, has_roo := parent_version["root"]
	if has_roo {
		root = roo.(bson.ObjectId)
	} else {
		root = parent
	}
	return
}

// At update uses $set, does not replace document.
func InudOpt(db *mgo.Database, ev ifaces.Event, dat map[string]interface{}, coll, op, id string, version bool) error {
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
		ins_id := bson.NewObjectId()
		version_id := bson.NewObjectId()
		dat["_id"] = ins_id
		// At an insert, the parent can be only a draft.
		dat["pointing_to"] = version_id
		draft_i, has_draft := dat["draft_id"]
		var parent, root bson.ObjectId
		if has_draft && len(draft_i.(string)) > 0 {
			draft_id := ToIdWithCare(draft_i.(string))
			parent, root, _, err = GetDraftParent(db, coll, draft_id)
			if err != nil {
				return err
			}
		}
		if parent == "" { // Parent is "" if we are not coming from a draft.
			dat["root"] = version_id
		} else {
			dat["root"] = root
		}
		err = db.C(coll).Insert(dat)
		if err != nil {
			return err
		}
		if version {
			err = SaveVersion(db, coll, version_id, ins_id, parent, root)
		}
	case "update":
		live_id := bson.ObjectIdHex(id)
		version_id := bson.NewObjectId()
		var parent, root bson.ObjectId
		var err error
		draft_i, has_draft := dat["draft_id"]
		if has_draft && len(draft_i.(string)) > 0 {
			draft_id := ToIdWithCare(draft_i.(string))
			parent, root, _, err = GetDraftParent(db, coll, draft_id)
		} else {
			parent, root, err = GetParentTroughContent(db, coll, live_id)
		}
		if err != nil {
			return err
		}
		q := bson.M{"_id": live_id}
		dat["pointing_to"] = version_id // Points to the new version now.
		if root != "" {
			dat["root"] = root
		}
		fmt.Println(dat)
		upd := bson.M{"$set": dat}
		err = db.C(coll).Update(q, upd)
		if err != nil {
			return err
		}
		if version {
			// Dat must contain -parent
			err = SaveVersion(db, coll, version_id, live_id, parent, root)
		}
	case "delete":
		live_id := bson.ObjectIdHex(id)
		err = Delete(db, coll, live_id)
	case "restore":
		// err = db.C(coll).Find
		// Not implemented yet.
	}
	if err != nil {
		return err
	}
	ev.Trigger(coll+"."+op, dat)
	return nil
}

// Converts all bson.M s to map[string]interface{} s. Usually called on db query results.
// Will become obsolete when the mgo driver will return map[string]interface{} maps instead of bson.M ones.
func Convert(x interface{}) interface{} {
	if y, ok := x.(bson.M); ok {
		for key, val := range y {
			y[key] = Convert(val)
		}
		return (map[string]interface{})(y)
	} else if d, ok := x.(map[string]interface{}); ok {
		for key, val := range d {
			d[key] = Convert(val)
		}
		return d
	} else if z, ok := x.([]interface{}); ok {
		for i, v := range z {
			z[i] = Convert(v)
		}
	}
	return x
}

// Creates a copy of the most up to date document in collname (sorted by sortfield), and returns it's ObjectId for further updates.
// Used in situations where we want to handle a series of documents as immutable values, like the the documents in the "options" collection.
func CreateCopy(db *mgo.Database, collname, sortfield string) bson.ObjectId {
	var v []interface{}
	err := db.C(collname).Find(nil).Sort(sortfield).Limit(1).All(&v)
	if err != nil {		// Refactor this into return value.
		panic(err)
	}
	var ma bson.M
	if len(v) == 0 {
		ma = bson.M{}
	} else {
		ma = v[0].(bson.M)
	}
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
func DateAndAuthor(rule map[string]interface{}, dat map[string]interface{}, user_id bson.ObjectId, updating bool) {
	inserting := !updating
	for i, _ := range rule {
		switch i {
		case Created_by:
			if inserting {
				dat[i] = user_id
			}
		case Last_modified_by:
			if updating {
				dat[i] = user_id
			}
		case Created:
			if inserting {
				dat[i] = time.Now().Unix()
			}
		case Last_modified:
			if updating {
				dat[i] = time.Now().Unix()
			}
		}
	}
}

// If rule has slugs, then dat will already contain it if it was sent from UI.
func Slug(rule map[string]interface{}, dat map[string]interface{}) {
	_, has_slug := rule["slug"]
	if has_slug {
		if slug, has := dat["slug"]; has && len(slug.(string)) > 0 {
			return
		} else if name, has_name := dat["name"]; has_name {
			dat["slug"] = slugify.S(name.(string))
		} else if title, has_title := dat["title"]; has_title {
			dat["slug"] = slugify.S(title.(string))
		}
	}
}

// From user interface the ObjectId can come in a form (OjbectIdHex("era3232322332dsds33")) which is inappropriate as an ObjectId input.
// This strips the unnecessary parts.
// Once the UI will be universally cleared from those unnecessary string parts it will become obsolete. (See IdsToStrings)
func StripId(str_id string) string {
	l := len(str_id)
	if l != 24 {
		if l < 38 {
			panic("Bad id at basic.StripId.")
		}
		return str_id[13:37]
	}
	return str_id
}

// Helps to extract a bunch of ids from the UI input.
func ExtractIds(dat map[string][]string, keys []string) ([]string, error) {
	ret := []string{}
	for _, v := range keys {
		id_s, has_id_s := dat[v]
		if !has_id_s {
			return nil, fmt.Errorf(v, " id is not found.")
		}
		id := id_s[0]
		if len(id) == 24 {
			ret = append(ret, id)
		} else if len(id) == 39 {
			ret = append(ret, id[13:37])
		} else {
			return nil, fmt.Errorf(id, " is not a proper id.")
		}
	}
	return ret, nil
}

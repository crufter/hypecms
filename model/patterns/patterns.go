// Package pattern contains (should, lol) reusable database patterns.
package patterns


import(
	"github.com/opesun/resolver"
	"github.com/opesun/hypecms/model/basic"
	"labix.org/v2/mgo/bson"
	"labix.org/v2/mgo"
	"fmt"
)

type m map[string]interface{}

func ToIdWithCare(id interface{}) bson.ObjectId {
	switch val := id.(type) {
	case bson.ObjectId:
	case string:
		id = bson.ObjectIdHex(basic.StripId(val))
	default:
		panic(fmt.Sprintf("Can't create bson.ObjectId out of %T", val))
	}
	return id.(bson.ObjectId)
}

// Finds a doc by field-value equality.
func FindEq(db *mgo.Database, coll, field string, value interface{}) (map[string]interface{}, error) {
	if field == "_id" {
		value = ToIdWithCare(value)
	}
	q := m{field: value}
	var res interface{}
	err := db.C(coll).Find(q).One(&res)
	if err != nil { return nil, err }
	if res == nil { return nil, fmt.Errorf("Can't find document at FindEq.") }
	doc := basic.Convert(res.(bson.M)).(map[string]interface{})
	return doc, nil
}

func FindChildren(db *mgo.Database, children_coll, parent_fk_field string, parent_id bson.ObjectId) ([]interface{}, error) {
	q := m{parent_fk_field: parent_id}
	var children []interface{}
	err := db.C(children_coll).Find(q).All(&children)
	if err != nil { return nil, err }
	if children == nil { return nil, fmt.Errorf("Can't find children.") }
	children = basic.Convert(children).([]interface{})
	dont_query := map[string]interface{}{"password":0}
	resolver.ResolveAll(db, children, dont_query)
	return children, nil
}

// Finds a document in [parent_coll] collection based on [field] [value] equality, then queries
// [children_coll] for documents which has the _id of that document in their parent_fk_field.
// Returns children list only, no parent.
func FindParentAndChildren(db *mgo.Database, parent_coll string, field string, value interface{}, children_coll, parent_fk_field string) ([]interface{}, error)  {
	parent, err := FindEq(db, parent_coll, field, value)
	if err != nil { return nil, err }
	return FindChildren(db, children_coll, parent_fk_field, parent["_id"].(bson.ObjectId))
}

func FieldStartsWith(db *mgo.Database, collname, fieldname, val string) ([]interface{}, error) {
	var res []interface{}
	q := m{fieldname: bson.RegEx{ "^" + val, "u"}}
	err := db.C(collname).Find(q).All(&res)
	if err != nil { return nil, err }
	if res == nil { return nil, fmt.Errorf("Can't find %v starting with %v.", fieldname, val) }
	res = basic.Convert(res).([]interface{})
	return res, nil
}

// Takes a collection, a field and a value and pulls that value from all docs in the collection.
// Caution, it does not take care about string id values, or even worse, non stripped string id values.
func PullFromAll(db *mgo.Database, collname, fieldname string, value interface{}) error {
	_, err := db.C(collname).UpdateAll(nil, m{"$pull":m{fieldname:value}})
	return err
}

// Removes a doc by id, takes care about non stripped string ids.
func DeleteById(db *mgo.Database, collname string, id interface{}) error {
	id = ToIdWithCare(id)
	q := m{"_id": id}
	return db.C(collname).Remove(q)
}

func IncAll(db *mgo.Database, collname string, ids []bson.ObjectId, fieldname string, num int) error {
	q := m{
		"_id": m{
			"$in": ids,
		},
	}
	upd := m{
		"$inc": m{
			fieldname: num,
		},
	}
	_, err := db.C(collname).UpdateAll(q, upd)
	return err
}

func Satisfies(db *mgo.Database, collname string, id bson.ObjectId, query map[string]interface{}) (bool, error) {
	query["_id"] = id
	count, err := db.C(collname).Find(query).Count()
	if err != nil { return false, err }
	if count == 1 { return true, nil }
	return false, fmt.Errorf("Does not satisfy.")
}
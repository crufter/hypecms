// Package pattern contains (should, lol) reusable database patterns.
// These functions are intended as inner building blocks of saner APIs.
package patterns


import(
	"github.com/opesun/hypecms/model/basic"
	"labix.org/v2/mgo/bson"
	"labix.org/v2/mgo"
	"fmt"
)

type m map[string]interface{}

func ToIdWithCare(id interface{}) bson.ObjectId {
	return basic.ToIdWithCare(id)
}

// Finds a doc by field-value equality.
func FindEq(db *mgo.Database, coll, field string, value interface{}) (map[string]interface{}, error) {
	q := m{field: value}
	return FindQ(db, coll, q)
}

// Finds a doc by query.
func FindQ(db *mgo.Database, coll string, query map[string]interface{}) (map[string]interface{}, error) {
	id, has := query["_id"]
	if has {
		query["_id"] = ToIdWithCare(id)
	}
	var res interface{}
	err := db.C(coll).Find(query).One(&res)
	if err != nil { return nil, err }
	if res == nil { return nil, fmt.Errorf("Can't find document at FindEq.") }
	doc := basic.Convert(res.(bson.M)).(map[string]interface{})
	return doc, nil
}

// See *1 below.
func FindChildren(db *mgo.Database, children_coll, parent_fk_field string, parent_id bson.ObjectId, additional_query map[string]interface{}) ([]interface{}, error) {
	q := map[string]interface{}{}
	if additional_query != nil {
		q = additional_query
	}
	q[parent_fk_field] = parent_id
	var children []interface{}
	err := db.C(children_coll).Find(q).All(&children)
	if err != nil { return nil, err }
	if children == nil { return nil, fmt.Errorf("Can't find children.") }
	children = basic.Convert(children).([]interface{})
	return children, nil
}

// [1] This is highly experimental, not used ATM, dont look here, its ugly.
//
// Finds a document in [parent_coll] collection based on [field] [value] equality, then queries
// [children_coll] for documents which has the _id of that document in their parent_fk_field.
// Returns children list only, no parent.
func FindChildrenByParent(db *mgo.Database, parent_coll string, parent_q map[string]interface{}, children_coll, parent_fk_field string, children_q map[string]interface{}) ([]interface{}, error)  {
	parent, err := FindQ(db, parent_coll, parent_q)
	if err != nil { return nil, err }
	return FindChildren(db, children_coll, parent_fk_field, parent["_id"].(bson.ObjectId), children_q)
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

func IncAll(db *mgo.Database, collname string, ids []bson.ObjectId, fieldnames []string, num int) error {
	q := m{
		"_id": m{
			"$in": ids,
		},
	}
	fields_to_inc := m{}
	for _, v := range fieldnames {
		fields_to_inc[v] = num
	}
	upd := m{
		"$inc": fields_to_inc,
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
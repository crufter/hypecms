package custom_actions_model

import(
	"labix.org/v2/mgo/bson"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/patterns"
	"github.com/opesun/hypecms/modules/content/model"
	"labix.org/v2/mgo"
	"fmt"
)

// With this two actions, one can implement a QA site, similar to StackOverflow.

// Example:
// {
//		"type":					"choose_child",
//		"c": 					"contents",
//		"choose_fieldname":		"accepted",			// This field will be set to true when the child document is chosen.
//		"parent_fieldname":		"parent",			// The name of the field in which the parent's id resides.
//		"max_choose_count":		1					// Optional, defaults to 1. Describes how many chosen children can exists per parent.
// }
func ChooseChild(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string) error {
	choose_fieldname 	:= action["choose_fieldname"].(string)
	parent_fieldname 	:= action["parent_fieldname"].(string)
	coll				:= action["c"].(string)
	max_choose_count 	:= 1
	mcc, has_mcc 		:= action["max_choose_count"]
	if has_mcc { max_choose_count = int(mcc.(float64)) }
	rule := m{
		"chosen_doc_id":	"must",
	}
	dat, err := extract.New(rule).Extract(inp)
	if err != nil { return err }
	chosen_doc_id := patterns.ToIdWithCare(dat["chosen_doc_id"])
	chosen_doc, err := patterns.FindEq(db, coll, "_id", chosen_doc_id)
	if err != nil { return err }
	parent_id := chosen_doc[parent_fieldname]
	// Check if author of equals to user_id, because only the parent author can chose a child.
	user_id := user["_id"].(bson.ObjectId)
	parent_doc, err := patterns.FindEq(db, coll, "_id", parent_id)
	if err != nil { return err }
	owner := parent_doc["_users_created_by"].(bson.ObjectId) == user_id
	if !owner { return fmt.Errorf("You can only choose a child of your own document.") }
	var res []interface{}
	// Find all chosen children.
	err = db.C(coll).Find(m{parent_fieldname: parent_id, choose_fieldname: true}).All(&res)
	if err != nil { return err }
	max_choose_exceeded := res != nil && len(res) > max_choose_count - 1
	if max_choose_exceeded { return fmt.Errorf("There is already %v chosen child(ren) of this parent document.", max_choose_count) }
	q := m{"_id": chosen_doc_id}
	upd := m{
		"$set": m{
			choose_fieldname: true,
		},
	}
	err = db.C(coll).Update(q, upd)
	if err != nil { return err }
	q = m{"_id": parent_id}
	upd = m{
		"$set": m{
			"has_" + choose_fieldname: true,
		},
	}
	return db.C(coll).Update(q, upd)
}

// RespondContent could be called "insert child content".
// Example:
// {
// 		"type": 					"respond",
//		"parent_fieldname":			"parent",
// }
//
// If the extract module could convert a string to bson.ObjectId one would not need this.
func RespondContent(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string, response_content_type_options map[string]interface{}) error {
	parent_fieldname := action["parent_fieldname"].(string)
	rcto := response_content_type_options
	parent, has_parent := inp["response_parent"]
	if !has_parent { return fmt.Errorf("No parent given.") }
	parent_id := patterns.ToIdWithCare(parent)
	fixval := m{parent_fieldname: parent_id}
	_, err := content_model.InsertWithFix(db, nil, rcto, inp, user["_id"].(bson.ObjectId), fixval)
	return err
}

// There is a problem with this: how to do update, delete?
// This update/delete problem does not exist at ResponContent, because the content module can do all of it.
// TODO: think about the possible solutions to this.
//
// Respond could be called "insert child"
// {
//		"c":		"responses",
//		"rules": {
//					"title": 1,
//					"content": 1,
//					"created": 1,
//					"_users_created_by": 1
//				}
// }
// func Respond(db *mgo.Database, user, action map[string]interrface{}, inp map[string]interface{}) error {
// }
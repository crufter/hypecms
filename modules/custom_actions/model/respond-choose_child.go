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

// Checks if user_id is the author of the parent of the chosen document.
func IsAuthor(db *mgo.Database, coll, parent_fieldname string, user_id, chosen_doc_id bson.ObjectId) (bson.ObjectId, error) {
	chosen_doc, err := patterns.FindEq(db, coll, "_id", chosen_doc_id)
	if err != nil { return "", err }
	parent_id := chosen_doc[parent_fieldname]
	// Check if author of equals to user_id, because only the parent author can chose a child.
	parent_doc, err := patterns.FindEq(db, coll, "_id", parent_id)
	if err != nil { return "", err }
	owner := parent_doc["_users_created_by"].(bson.ObjectId) == user_id
	if !owner { return "", fmt.Errorf("You can only choose a child of your own document.") }
	return parent_id.(bson.ObjectId), nil
}

// TODO: This could be done by checking the value of "has_" + choose_fieldname in parent.
func Exceeded(db *mgo.Database, coll, parent_fieldname, choose_fieldname string, parent_id bson.ObjectId, max_choices int) (int, error) {
	var res []interface{}
	// Find all chosen children.
	err := db.C(coll).Find(m{parent_fieldname: parent_id, choose_fieldname: true}).All(&res)
	if err != nil { return 0, err }
	current_count := len(res)
	max_choices_exceeded := res != nil && current_count > max_choices - 1
	if max_choices_exceeded { return 0, fmt.Errorf("There is already %v chosen child(ren) of this parent document.", max_choices) }
	return current_count, nil
}

func MarkAsChosen(db *mgo.Database, coll, choose_fieldname string, chosen_doc_id bson.ObjectId, current_count int) error {
	q := m{"_id": chosen_doc_id}
	upd := m{
		"$set": m{
			choose_fieldname: current_count + 1,
		},
	}
	return db.C(coll).Update(q, upd)
}

func UnmarkAsChosen(db *mgo.Database, coll, choose_fieldname string, chosen_doc_id bson.ObjectId) error {
	q := m{"_id": chosen_doc_id}
	upd := m{
		"$unset": m{
			choose_fieldname: 1,
		},
	}
	return db.C(coll).Update(q, upd)
}

func IncrementParent(db *mgo.Database, coll, choose_fieldname string, parent_id bson.ObjectId, by int) error {
	q := m{"_id": parent_id}
	upd := m{
		"$inc": m{
			"has_" + choose_fieldname: by,
		},
	}
	return db.C(coll).Update(q, upd)
}

// Example:
// {
//		"type":					"choose_child",
//		"c": 					"contents",
//		"choose_fieldname":		"accepted",			// This field will be set to true when the child document is chosen.
//		"parent_fieldname":		"parent",			// The name of the field in which the parent's id resides.
//		"max_choices":			1					// Optional, defaults to 1. Describes how many chosen children can exists per parent.
// }
//
// Gets chosen document to find out parent.
// Gets parent document, checks if user is the author of it. Return error if not.
// Sets choose_fieldname to the number of already chosen children + 1 in the chosen document.
// Increases "has_" + choose_fieldname by 1 in parent document.
func ChooseChild(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string) error {
	choose_fieldname 	:= action["choose_fieldname"].(string)
	parent_fieldname 	:= action["parent_fieldname"].(string)
	coll				:= action["c"].(string)
	max_choices		 	:= 1
	mc, has_mc 			:= action["max_choices"]
	if has_mc { max_choices = int(mc.(float64)) }
	rule := m{
		"chosen_doc_id":	"must",
	}
	dat, err := extract.New(rule).Extract(inp)
	if err != nil { return err }
	user_id := user["_id"].(bson.ObjectId)
	chosen_doc_id := patterns.ToIdWithCare(dat["chosen_doc_id"])
	parent_id, auth_err := IsAuthor(db, coll, parent_fieldname, user_id, chosen_doc_id)
	if auth_err != nil { return err }
	current_count, exc_err := Exceeded(db, coll, parent_fieldname, choose_fieldname, parent_id, max_choices)
	if exc_err != nil { return err }
	err = MarkAsChosen(db, coll, choose_fieldname, chosen_doc_id, current_count)
	if err != nil { return err }
	return IncrementParent(db, coll, choose_fieldname, parent_id, 1)
}

// Gets chosen document to find out parent.
// Gets parent document, checks if user is the author of it. Return error if not.
// Unsets choose_fieldname in chosen document.
// Decreases "has_" + choose_fieldname by 1 in parent document.
func UnchooseChild(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string) error {
	choose_fieldname 	:= action["choose_fieldname"].(string)
	parent_fieldname 	:= action["parent_fieldname"].(string)
	coll				:= action["c"].(string)
	return nil
	rule := m{
		"chosen_doc_id":	"must",
	}
	dat, err := extract.New(rule).Extract(inp)
	if err != nil { return err }
	user_id := user["_id"].(bson.ObjectId)
	chosen_doc_id := patterns.ToIdWithCare(dat["chosen_doc_id"])
	parent_id, auth_err := IsAuthor(db, coll, parent_fieldname, user_id, chosen_doc_id)
	if auth_err != nil { return auth_err }
	err = UnmarkAsChosen(db, coll, choose_fieldname, chosen_doc_id)
	if err != nil { return err }
	return IncrementParent(db, coll, choose_fieldname, parent_id, -1)
}

// RespondContent could be called "insert child content".
// Example:
// {
// 		"type": 									"respond",
//		"parent_fieldname":							"parent",
//		"counter_fieldname":						"replies",
//		"can_delete":								true,								// Optional, defaults to false
//		"delete_unless":							{"$exists": "accepted"}				// Optional, possible to delete response, unless this query on the response evaluates to true.
//		"update_parent_on_delete_if_child":			[[{"$exists":"accepted"},			// Optional, run this update query on parent.
//													{"$unset": {"has_accepted": 1}}], ...]
//		
// }
//
// If the extract module could convert a string to bson.ObjectId we would not need this.
//
// Checks if parent exists. Returns error if not.
// Inserts new content with the id of the parent.
// Increments field named "counter_fieldname" in parent.
func RespondContent(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string, response_content_type_options map[string]interface{}) error {
	parent_fieldname 	:= action["parent_fieldname"].(string)
	counter_fieldname 	:= action["counter_fieldname"].(string)
	rcto := response_content_type_options
	parent, has_parent := inp["response_parent"]
	if !has_parent { return fmt.Errorf("No parent given.") }
	parent_id := patterns.ToIdWithCare(parent)
	q := m{"_id": parent_id}
	count, err := db.C("contents").Find(q).Count()
	if err != nil { return err }
	if count == 0 { return fmt.Errorf("Can't find parent with id %v.", parent_id) }
	fixval := m{parent_fieldname: parent_id}
	_, err = content_model.InsertWithFix(db, nil, rcto, inp, user["_id"].(bson.ObjectId), fixval)
	if err != nil { return err }
	upd := m{
		"$inc": m{
			counter_fieldname: 1,
		},
	}
	return db.C("contents").Update(q, upd)
}

// We must somehow forbid the deletion of a given ContentResponse, because we have to do these actions at deletion too.
// TODO: we have to find a solution to this question.
// Update has no problems, no additional processing needs to be done when updating some fields like title, content, etc.
// See https://github.com/opesun/hypecms/issues/29
// (Or operate with events/hooks?)
//
// Checks if it is possible to delete a response.
// Checks if the response satisfies the "delete_unless" query, returns error if it does.
// Deletes response.
// Decrements parent "counter_fieldname" field by one.
// Does "update_parent_on_delete_if_child" optional operation.
func DeleteContentResponse(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string) error {
	can_delete := false
	if can_delete_i, has := action["can_delete"]; has {
		can_delete = can_delete_i.(bool)
	}
	if !can_delete { return fmt.Errorf("Can't delete response.") }
	response_id :=	patterns.ToIdWithCare(inp["respone_id"][0])
	delete_unless, has_du := action["delete_unless"]
	if has_du {
		sat, err := patterns.Satisfies(db, "contents", response_id, delete_unless.(map[string]interface{}))
		if err != nil { return err }
		if sat {
			return fmt.Errorf("Can't delete response, satisfies \"unless\" query.")
		}
	}
	//counter_fieldname 	:= action["counter_fieldname"].(string)
	//parent_update, has_pu := action["update_parent"]
	return nil
}

// There is a problem with this: how to do update, delete?
// This update/delete problem does not exist at ResponContent, because the content module can do all of it.
// TODO: think about the possible solutions to this.
//
// Respond could be called "insert child document"
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
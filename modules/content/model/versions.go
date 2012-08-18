package content_model

import(
	ifaces "github.com/opesun/hypecms/interfaces"
	"labix.org/v2/mgo"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/hypecms/model/patterns"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/extract"
	"fmt"
	"time"
)

const(
	Draft_collection_postfix 	= "_draft"
	Parent_content_field		= "parent_content"
	Parent_draft_field			= "parent_draft"
	no_right_draft 				= "You have no rights to save a draft of type %v."
)

func compLev(req_lev, user_level int, typ string) error {
	if user_level >= req_lev { return nil }
	return fmt.Errorf(no_right_draft, typ)
}

// content_type_options: Modules.content.types.[type]
func AllowsDraft(content_type_options map[string]interface{}, user_level int, typ string) error {
	req_lev := 300
	val, has := content_type_options["draft_level"]
	if !has { return compLev(req_lev, user_level, typ) }
	req_lev = int(val.(int64))
	return compLev(req_lev, user_level, typ)
}

// Implementation of versioning is in basic.InudVersion.
func RevertToVersion(db *mgo.Database, ev ifaces.Event, inp map[string][]string, non_versioned_fields []string) error {
	rule := map[string]interface{}{
		"id": "must",
	}
	dat, err := extract.New(rule).Extract(inp)
	if err != nil { return err }
	id_str := basic.StripId(dat["_id"].(string))
	id := bson.ObjectIdHex(id_str)
	var v interface{}
	err = db.C(Cname).Find(bson.M{"_id": id}).One(&v)
	old_version	:= v.(bson.M)
	if err != nil { return err }
	if v == nil { return fmt.Errorf("Can't find content at RevertToVersion.") }
	parent_id := old_version[basic.Version_parentfield].(bson.ObjectId)
	for _, v := range non_versioned_fields {
		delete(old_version, v)
	}
	old_version["_contents_reverted_to"] = id
	return basic.InudVersion(db, ev, old_version, Cname, "update", parent_id.Hex())
}

// We never update drafts, we always insert a new one.
// A draft will have the next fields: id, type, created, up_to_date, parent_content/draft_content/none/both, data.
// The saved input resides in the data.
func SaveDraft(db *mgo.Database, content_rules map[string]interface{}, inp map[string][]string) (bson.ObjectId, error) {
	t := time.Now()
	for i, _ := range content_rules {
		content_rules[i] = 1
	}
	content_rules["type"] 						= 	"must"
	content_rules[Parent_content_field] 		= 	"must"
	content_rules[Parent_draft_field]			= 	"must"
	dat, err := extract.New(content_rules).Extract(inp)
	parent_content_id_str := dat[Parent_content_field].(string)
	parent_draft_id_str := dat[Parent_draft_field].(string)
	delete(dat, Parent_content_field)
	delete(dat, Parent_draft_field)
	if err != nil { return "", err }
	ins := m{
		"created":		time.Now().Unix(),
		"data":			dat,
		"type":			dat["type"].(string) + "_draft",
		"up_to_date":	true,
	}
	if len(parent_content_id_str) > 0 {
		parent_id := patterns.ToIdWithCare(parent_content_id_str)
		ins[Parent_content_field] = parent_id
	}
	if len(parent_draft_id_str) > 0 {
		parent_draft_id := patterns.ToIdWithCare(parent_draft_id_str)
		q := m{"_id": parent_draft_id}
		upd := m{
			"$unset": m{
				"up_to_date":	1,
			},
		}
		err = db.C(Cname + Draft_collection_postfix).Update(q, upd)
		if err != nil { return "", err }	// Rollback previous update.
		ins[Parent_draft_field] = parent_draft_id
	}
	draft_id := bson.NewObjectId()
	ins["_id"] = draft_id
	ins["kind"] = "draft"
	err = db.C(Cname + Draft_collection_postfix).Insert(ins)
	fmt.Println(time.Since(t))
	return draft_id, err
}

// draft["data"] will contain draft["data"] merged with all of the parent's fields.
func mergeWithParent(dat, parent map[string]interface{}) map[string]interface{} {
	for i, v := range dat {
		parent[i] = v
	}
	return parent
}

// Parent can be content/draft/both/none.
func HasContentParent(draft map[string]interface{}) bool {
	if parent_content_i, has_parent_cont := draft[Parent_content_field]; has_parent_cont {
		_, ok := parent_content_i.(bson.ObjectId)
		if !ok { panic("\"parent_content\" field is not an instance of bson.ObjectId.") }
		return true
	}
	return false
}

// Parent can be content or draft.
func HasDraftParent(draft map[string]interface{}) bool {
	if parent_draft_i, has_parent_draft := draft[Parent_draft_field]; has_parent_draft {
		_, ok := parent_draft_i.(bson.ObjectId)
		if !ok { panic("\"parent_draft\" field is not an instance of bson.ObjectId.") }
		return true
	}
	return false
}

// It's possible that it has no parent at all, then it is a fresh new draft, first version.
func HasNoParent(draft map[string]interface{}) bool {
	if !HasContentParent(draft) && !HasDraftParent(draft) {
		return true
	}
	return false
}

// Queries a draft and rebuilds it. Queries its parent too, and merges it with the input fields saved in "data".
// The returned draft will be like a simple draft in the database, but in the data field it will contain fields of the parent plus the fresher saved input data.
// draft_typ example: blog_draft
func BuildDraft(db *mgo.Database, draft_typ, draft_id_str string) (map[string]interface{}, error) {
	draft_id := patterns.ToIdWithCare(draft_id_str)
	q := m{"_id": draft_id}
	var v interface{}
	err := db.C(Cname + Draft_collection_postfix).Find(q).One(&v)
	if err != nil { return nil, err }
	if v == nil { return nil, fmt.Errorf("Can't find draft.") }
	draft := basic.Convert(v).(map[string]interface{})
	typ_i, has_typ := draft["type"]
	if !has_typ { return nil, fmt.Errorf("Draft has no type.") }
	typ, is_str := typ_i.(string)
	if !is_str { return nil, fmt.Errorf("Draft type is not a string.") }
	if typ != draft_typ { return nil, fmt.Errorf("Draft type is not the expected one: %v instead if %v.", typ, draft_typ) }
	var parent_id bson.ObjectId
	if HasContentParent(draft) {
		parent_id = draft[Parent_content_field].(bson.ObjectId)
	}
	var par interface{}
	q = m{"_id": parent_id}
	err = db.C(Cname).Find(q).One(&par)
	parent := map[string]interface{}{}
	if err == nil || par != nil {
		parent = basic.Convert(par).(map[string]interface{})
	}
	draft["data"] = mergeWithParent(draft["data"].(map[string]interface{}), parent)
	return draft, nil
}

// Takes a content list and connects an up to date draft to each content if it exists.
// A draft is up to date if it is newer than the last saved comment, and newer than any other draft.
func ConnectWithDrafts(db *mgo.Database, content_list []interface{}) error {
	ids := []bson.ObjectId{}
	cache := map[string]int{}
	for i, doc_i := range content_list {
		doc := doc_i.(map[string]interface{})
		id := doc["_id"].(bson.ObjectId)
		cache[id.Hex()] = i
		ids = append(ids, id)
	}
	q := m{Parent_content_field: m{"$in": ids}, "up_to_date":true}
	var drafts []interface{}
	err := db.C(Cname + Draft_collection_postfix).Find(q).All(&drafts)
	if err != nil { return err }
	drafts = basic.Convert(drafts).([]interface{})
	for _, v := range drafts {
		draft := v.(map[string]interface{})
		if cont_par, has_cont_par := draft[Parent_content_field]; has_cont_par {
			cont_par_id := cont_par.(bson.ObjectId)
			subj_ind := cache[cont_par_id.Hex()]
			content := content_list[subj_ind].(map[string]interface{})
			content_last_mod, has_last_mod := content[basic.Last_modified]
			if !has_last_mod {
				content["latest_draft"] = v		// This will be shown errorenously if the content will never have a last modified field.
				continue
			}
			if content_last_mod.(int64) < draft[basic.Created].(int64) {
				content["latest_draft"] = v
			}
		}
	}
	return nil
}

// Takes a content id and gives back the up to date draft of it if it has one.
// Gives back nil if it does not have one.
func GetUpToDateDraft(db *mgo.Database, content_id bson.ObjectId, content map[string]interface{}) map[string]interface{} {
	var latest_draft interface{}
	db.C(Cname + Draft_collection_postfix).Find(m{Parent_content_field: content_id, "up_to_date": true}).One(&latest_draft)
	if latest_draft == nil { return nil }
	draft := latest_draft.(bson.M)
	content_last_mod, has_last_mod := content[basic.Last_modified]
	if !has_last_mod { return draft }
	if content_last_mod.(int64) < draft[basic.Created].(int64) {
		return draft
	}
	return nil
}

// Takes a draft, and a parent content and decides if the draft is up to date or not.
func IsDraftUpToDate(db *mgo.Database, draft, parent map[string]interface{}) (bool, error) {
	parent_last_mod, has_last_mod := parent[basic.Last_modified]
	if !has_last_mod { return true, nil }
	fresher_than_parent := parent_last_mod.(int64) < draft[basic.Created].(int64)
	if !fresher_than_parent { return false, nil }
	var v interface{}
	q := m{Parent_content_field: draft[Parent_content_field], "up_to_date": true} 
	err := db.C(Cname + Draft_collection_postfix).Find(q).One(&v)	// TODO: no error checking here.
	if err != nil { return false, err }
	if v == nil { return false, fmt.Errorf("Can't find any draft at IsDraftUpToDate.") }
	if v.(bson.M)["_id"].(bson.ObjectId) != draft["_id"].(bson.ObjectId) { return false, nil }
	return true, nil
}
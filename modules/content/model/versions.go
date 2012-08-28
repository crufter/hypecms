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
	Parent_content_field		= "draft_of"
	Parent_draft_field			= "draft_id"
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
func ChangeHead(db *mgo.Database, ev ifaces.Event, inp map[string][]string, non_versioned_fields []string) error {
	rule := map[string]interface{}{
		"version_id": 	"must",
		"id":			"must",
	}
	dat, err := extract.New(rule).Extract(inp)
	if err != nil { return err }
	version_id_str := basic.StripId(dat["version_id"].(string))
	version_id := bson.ObjectIdHex(version_id_str)
	var v interface{}
	err = db.C(Cname + "_version").Find(bson.M{"_id": version_id}).One(&v)
	if err != nil { return err }
	revert_to	:= v.(bson.M)
	id := patterns.ToIdWithCare(dat["id"].(string))
	for _, v := range non_versioned_fields {
		delete(revert_to, v)
	}
	revert_to["points_to"] = revert_to["_id"]
	delete(revert_to, "id")
	return db.C(Cname).Update(bson.M{"_id": id}, bson.M{"$set": revert_to})
}

// We never update drafts, we always insert a new one.
// A draft will have the next fields: id, type, created, up_to_date, parent_content/draft_content/none/both, data.
// The saved input resides in the data.
func SaveDraft(db *mgo.Database, content_rules map[string]interface{}, inp map[string][]string) (bson.ObjectId, error) {
	for i, _ := range content_rules {
		content_rules[i] = 1
	}
	content_rules["type"]						=	"must"
	content_rules["id"] 						= 	"must"		// Content id.
	content_rules["draft_id"] 					= 	"must"		// Draft id, no empty if we save draft from a draft, empty if draft is saved from a content.
	dat, err := extract.New(content_rules).Extract(inp)
	if err != nil { return "", err }
	live_id_s := dat["id"].(string)
	draft_id_s := dat["draft_id"].(string)
	ins := m{
		"created":		time.Now().Unix(),
		"data":			dat,
		"type":			dat["type"],
	}
	var parent, root bson.ObjectId
	if len(draft_id_s) > 0 {	// Coming from a draft.
		draft_id := patterns.ToIdWithCare(draft_id_s)
		var last_version bson.ObjectId
		parent, root, last_version, err = basic.GetDraftParent(db, "contents", draft_id)
		if parent == "" || root == "" { return "", fmt.Errorf("Coming from draft but still can not extract parent or root.") }
		// Has no version of it is a draft/descendant of a draft which connects to no saved content.
		if last_version != "" {
			ins["draft_of_version"] = last_version
		}
	} else if len(live_id_s) > 0 {	// Coming from a content.
		live_id := patterns.ToIdWithCare(live_id_s)
		// This way draft saves coming from any versions will be saved to the version pointed to by the head anyway...
		parent, root, err = basic.GetParentTroughContent(db, "contents", live_id)
		// If we have a live content saved, we must have a version to point to too.
		if parent == "" || root == "" { return "", fmt.Errorf("Coming from content but still can not extract parent or root.") }
		ins["draft_of_version"] = parent
	}
	if err != nil { return "", err }
	if len(live_id_s) > 0 {
		live_id := patterns.ToIdWithCare(live_id_s)
		ins["draft_of"] = live_id		// Just so we can display the content id for saving content immediately from a draft.
	}
	draft_id := bson.NewObjectId()
	if parent != "" {	// If there is a parent there is a root...
		ins["-parent"] = parent
		ins["root"] = root
	} else {
		ins["root"] = draft_id
	}
	// Note: we dont store the last version here, building the draft will be harder.
	ins["_id"] = draft_id
	err = db.C(Cname + Draft_collection_postfix).Insert(ins)
	return draft_id, err
}

// draft["data"] will contain draft["data"] merged with all of the parent's fields.
func mergeWithParent(dat, parent map[string]interface{}) map[string]interface{} {
	if dat == nil { return parent }
	if parent == nil { return dat }
	for i, v := range dat {
		parent[i] = v
	}
	return parent
}

// Takes a draft and gets.
func GetParent(db *mgo.Database, coll string, draft map[string]interface{}) (map[string]interface{}, error) {
	version_parent, has_vp := draft["draft_of_version"]
	_, has_draft_of := draft["draft_of"]
	// If a draft is a draft of an existing content, than it must have a parent version.
	if !has_vp && has_draft_of {
		return nil, fmt.Errorf("State of draft is inconsistent, parent version is not set.")
	}
	// Simply the draft is not connected to anything saved.
	if !has_vp {	// && !has_draft_of
		return nil, nil
	}
	var par interface{}
	parent_id := version_parent.(bson.ObjectId)
	q := m{"_id": parent_id}
	err := db.C(Cname + "_version").Find(q).One(&par)
	if err != nil { return nil, err }
	return basic.Convert(par).(map[string]interface{}), nil
}

// Queries a draft and rebuilds it. Queries its parent too, and merges it with the input fields saved in "data".
// The returned draft will be like a simple draft in the database, but in the data field it will contain fields of the parent plus the fresher saved input data.
// draft_typ example: blog_draft
// Only
func BuildDraft(db *mgo.Database, draft_typ, draft_id_str string) (map[string]interface{}, error) {
	draft_id := patterns.ToIdWithCare(draft_id_str)
	q := m{"_id": draft_id}
	var v interface{}
	err := db.C(Cname + Draft_collection_postfix).Find(q).One(&v)
	if err != nil { return nil, err }
	draft := basic.Convert(v).(map[string]interface{})
	typ_i, has_typ := draft["type"]
	if !has_typ { return nil, fmt.Errorf("Draft has no type.") }
	typ, is_str := typ_i.(string)
	if !is_str { return nil, fmt.Errorf("Draft type is not a string.") }
	if typ != draft_typ { return nil, fmt.Errorf("Draft type is not the expected one: %v instead if %v.", typ, draft_typ) }
	parent, err := GetParent(db, "contents", draft)
	if err != nil { return nil, err }
	draft["data"] = mergeWithParent(draft["data"].(map[string]interface{}), parent)
	return draft, nil
}

// Takes a content list and connects an up to date draft to each content if it exists.
// A draft is up to date if it is newer than the last saved comment, and newer than any other draft.
func HaveUpToDateDrafts(db *mgo.Database, content_list []interface{}) error {
	return nil
}

// Takes a content id and gives back the up to date draft of it if it has one.
// Gives back nil if it does not have one.
func GetUpToDateDraft(db *mgo.Database, content_id bson.ObjectId, content map[string]interface{}) map[string]interface{} {
	return nil
}

// Takes a draft, and a parent content and decides if the draft is up to date or not.
func IsDraftUpToDate(db *mgo.Database, draft, parent map[string]interface{}) (bool, error) {
	return true, nil
}

// See *1
func DraftTimeline(db *mgo.Database, draft_id bson.ObjectId) ([]interface{}, error) {
	var x interface{}
	q := m{"_id": draft_id}
	err := db.C(Cname + Draft_collection_postfix).Find(q).One(&x)
	if err != nil { return nil, err}
	draft_doc := x.(bson.M)
	root, has_root := draft_doc["root"]
	if !has_root {
		return []interface{}{}, nil
	}
	return GetFamily(db, root.(bson.ObjectId))
}

func GetFamily(db *mgo.Database, root bson.ObjectId) ([]interface{}, error) {
	ret := []interface{}{}
	var v []interface{}
	q := m{"root": root}
	err := db.C(Cname + Draft_collection_postfix).Find(q).Sort("created").All(&v)
	if err != nil { return nil, err }
	if v != nil {
		ret = append(ret, v...)
	}
	var r []interface{}
	err = db.C(Cname + "_version").Find(q).Sort("version_date").All(&r)
	if err != nil { return nil, err }
	if r != nil {
		ret = append(ret, r...)
	}
	var c []interface{}
	err = db.C(Cname).Find(q).All(&c)
	if err != nil { return nil, err }
	if c != nil {
		ret = append(ret, c...)
	} else if r != nil {		// If we have versions we should have a head too.
		return nil, fmt.Errorf("Inconsistent state in db: no head, but we have versions.")
	}
	return ret, nil
}

// *1. Temporary solution. With helper collections we could do the sorting with mongodb. Unfortunately, now versions and drafts reside in two separate collections.
func ContentTimeline(db *mgo.Database, content_doc map[string]interface{}) ([]interface{}, error) {
	root, has_r := content_doc["root"]
	if !has_r {
		return nil, fmt.Errorf("A content must have a root at any time.")
	}
	re, err := GetFamily(db, root.(bson.ObjectId))
	return re, err
}

// Find version by id.
func FindVersion(db *mgo.Database, id bson.ObjectId) (map[string]interface{}, error) {
	var v interface{}
	err := db.C("contents_version").Find(m{"_id": id}).One(&v)
	if err != nil { return nil, err }
	return basic.Convert(v).(map[string]interface{}), err
}
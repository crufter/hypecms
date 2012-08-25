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
	fmt.Println("started to save draft")
	if err != nil { return "", err }
	live_id_s := dat["id"].(string)
	draft_id_s := dat["draft_id"].(string)
	ins := m{
		"created":		time.Now().Unix(),
		"data":			dat,
		"type":			dat["type"],
	}
	var parent, root bson.ObjectId
	if len(draft_id_s) > 0 {
		draft_id := patterns.ToIdWithCare(draft_id_s)
		parent, root, err = basic.GetDraftParent(db, "contents", draft_id)
	} else if len(live_id_s) > 0 {
		// There should be a traversal of the tree upward to check if the head is not on a different path.
		live_id := patterns.ToIdWithCare(live_id_s)
		parent, root, err = basic.GetParentTroughContent(db, "contents", live_id)
	}
	if err != nil { return "", err }
	if len(live_id_s) > 0 {
		live_id := patterns.ToIdWithCare(live_id_s)
		ins["draft_of"] = live_id		// Just so we can display the content id for saving content immediately from a draft.
	}
	if parent != "" {	// If there is a parent there is a root...
		ins["-parent"] = parent
		ins["root"] = root
	}
	// Note: we dont store the last version here, building the draft will be harder.
	draft_id := bson.NewObjectId()
	ins["_id"] = draft_id
	fmt.Println(ins, err)
	err = db.C(Cname + Draft_collection_postfix).Insert(ins)
	fmt.Println("errrrrrrrrrrrrrrrrr", err, draft_id)
	return draft_id, err
}

// draft["data"] will contain draft["data"] merged with all of the parent's fields.
func mergeWithParent(dat, parent map[string]interface{}) map[string]interface{} {
	for i, v := range dat {
		parent[i] = v
	}
	return parent
}

// Takes a draft and queries all the version it stems from.
func GetParent(db *mgo.Database, coll string, draft map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("draft at getp:", draft)
	par_i, has_par := draft["-parent"]
	_, has_draft_of := draft["draft_of"]
	fmt.Println("wow", par_i, has_par, has_draft_of)
	// If a draft is a draft of an existing content, than it must have a parent (if not else, the version which the existing content points to).
	if !has_par && has_draft_of {
		return nil, fmt.Errorf("State of draft is inconsistent, parent is not set.")
	}
	// There is no parent at all, return and go on.
	if !has_par {	// && !has_draft_of
		return nil, nil
	}
	if has_par && !has_draft_of {	// Perfectly valid, draft has parent draft, but not connected to anything saved.
		return nil, nil
	}
	var p map[string]interface{}
	parent_id := par_i.(bson.ObjectId)
	search_coll := "_draft"
	counter := 0
	for {
		var par []interface{}
		q := m{"_id": parent_id}
		err := db.C(Cname + search_coll).Find(q).All(&par)
		if err != nil { return nil, err }
		if len(par) == 0 {
			if search_coll == "_draft" {
				search_coll = "_version"
			} else {
				return nil, fmt.Errorf("Can't find version.")
			}
		} else if len(par) == 1 {
			if search_coll == "_draft" {
				parent_id = par[0].(bson.M)["-parent"].(bson.ObjectId)
			} else {
				p = basic.Convert(par[0].(bson.M)).(map[string]interface{})
				break
			}
		} else {
			return nil, fmt.Errorf("Wtf.")
		}
		counter++
		if counter > 500 { panic("Wtf.") }
	}
	return p, nil
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
func ConnectWithDrafts(db *mgo.Database, content_list []interface{}) error {
	//ids := []bson.ObjectId{}
	//cache := map[string]int{}
	//for i, doc_i := range content_list {
	//	doc := doc_i.(map[string]interface{})
	//	id := doc["_id"].(bson.ObjectId)
	//	cache[id.Hex()] = i
	//	ids = append(ids, id)
	//}
	//q := m{Parent_content_field: m{"$in": ids}, "up_to_date":true}
	//var drafts []interface{}
	//err := db.C(Cname + Draft_collection_postfix).Find(q).All(&drafts)
	//if err != nil { return err }
	//drafts = basic.Convert(drafts).([]interface{})
	//for _, v := range drafts {
	//	draft := v.(map[string]interface{})
	//	if cont_par, has_cont_par := draft[Parent_content_field]; has_cont_par {
	//		cont_par_id := cont_par.(bson.ObjectId)
	//		subj_ind := cache[cont_par_id.Hex()]
	//		content := content_list[subj_ind].(map[string]interface{})
	//		content_last_mod, has_last_mod := content[basic.Last_modified]
	//		if !has_last_mod {
	//			content["latest_draft"] = v		// This will be shown errorenously if the content will never have a last modified field.
	//			continue
	//		}
	//		if content_last_mod.(int64) < draft[basic.Created].(int64) {
	//			content["latest_draft"] = v
	//		}
	//	}
	//}
	return nil
}

// Takes a content id and gives back the up to date draft of it if it has one.
// Gives back nil if it does not have one.
func GetUpToDateDraft(db *mgo.Database, content_id bson.ObjectId, content map[string]interface{}) map[string]interface{} {
	//var latest_draft interface{}
	//db.C(Cname + Draft_collection_postfix).Find(m{Parent_content_field: content_id, "up_to_date": true}).One(&latest_draft)
	//if latest_draft == nil { return nil }
	//draft := latest_draft.(bson.M)
	//content_last_mod, has_last_mod := content[basic.Last_modified]
	//if !has_last_mod { return draft }
	//if content_last_mod.(int64) < draft[basic.Created].(int64) {
	//	return draft
	//}
	return nil
}

// Takes a draft, and a parent content and decides if the draft is up to date or not.
func IsDraftUpToDate(db *mgo.Database, draft, parent map[string]interface{}) (bool, error) {
	//parent_last_mod, has_last_mod := parent[basic.Last_modified]
	//if !has_last_mod { return true, nil }
	//fresher_than_parent := parent_last_mod.(int64) < draft[basic.Created].(int64)
	//if !fresher_than_parent { return false, nil }
	//var v interface{}
	//q := m{Parent_content_field: draft[Parent_content_field], "up_to_date": true} 
	//err := db.C(Cname + Draft_collection_postfix).Find(q).One(&v)	// Sorting is just for safety purposes.
	//if err != nil { return false, err }
	//if v == nil { return false, fmt.Errorf("Can't find any draft at IsDraftUpToDate.") }
	//if v.(bson.M)["_id"].(bson.ObjectId) != draft["_id"].(bson.ObjectId) { return false, nil }
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
	err := db.C(Cname + Draft_collection_postfix).Find(q).All(&v)
	if err != nil { return nil, err }
	if v != nil {
		ret = append(ret, v...)
	}
	var r []interface{}
	err = db.C(Cname + "_version").Find(q).All(&r)
	if err != nil { return nil, err }
	if r != nil {
		ret = append(ret, r...)
	}
	var c []interface{}
	err = db.C(Cname).Find(q).All(&c)
	if err != nil { return nil, err }
	if c != nil {
		ret = append(ret, c...)
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
	fmt.Println("::::", re, err)
	return re, err
}
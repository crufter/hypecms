package content_model

import(
	ifaces "github.com/opesun/hypecms/interfaces"
	"labix.org/v2/mgo"
	"github.com/opesun/hypecms/model/basic"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/extract"
	"fmt"
	"time"
	"strings"
)

func RevertToVersion(db *mgo.Database, ev ifaces.Event, inp map[string][]string, non_versioned_fields []string) error {
	rule := map[string]interface{}{
		"id": "must",
	}
	dat, err := extract.New(rule).Extract(inp)
	if err != nil { return err }
	id_str := basic.StripId(dat["_id"].(string))
	id := bson.ObjectIdHex(id_str)
	var v interface{}
	err = db.C("contents").Find(bson.M{"_id": id}).One(&v)
	old_version	:= v.(bson.M)
	if err != nil { return err }
	if v == nil { return fmt.Errorf("Can't find content at RevertToVersion.") }
	parent_id := old_version[basic.Version_parentfield].(bson.ObjectId)
	for _, v := range non_versioned_fields {
		delete(old_version, v)
	}
	return basic.InudVersion(db, ev, old_version, "contents", "update", parent_id.Hex())
}

// We never update drafts.
func SaveDraft(db *mgo.Database, content_rules map[string]interface{}, inp map[string][]string) (bson.ObjectId, error) {
	for i, _ := range content_rules {
		content_rules[i] = 1
	}
	content_rules["id"] 		= 	"must"		// Needed?
	content_rules["type"] 		= 	"must"
	content_rules["prev_draft"]	= 	"must"
	dat, err := extract.New(content_rules).Extract(inp)
	if err != nil { return "", err }
	ins := m{
		"created":		time.Now().Unix(),
		"data":			dat,
		"type":			dat["type"].(string) + "_draft",
		"up_to_date":	true,
	}
	id_str := dat["id"].(string)
	if len(id_str) > 0 {
		parent_id := bson.ObjectIdHex(basic.StripId(id_str))
		ins["parent_content"] = parent_id
	} else {
		ins["parent_content"] = nil
	}
	parent_draft_id_str := dat["prev_draft"].(string)
	if len(parent_draft_id_str) > 0 {
		parent_draft_id := bson.ObjectIdHex(basic.StripId(parent_draft_id_str))
		q := m{"_id": parent_draft_id}
		upd := m{
			"$unset": m{
				"up_to_date":	1,
			},
		}
		err = db.C("contents").Update(q, upd)
		if err != nil { return "", err }	// Rollback previous update.
		ins["parent_draft"] = parent_draft_id
	}
	draft_id := bson.NewObjectId()
	ins["_id"] = draft_id
	return draft_id, db.C("contents").Insert(ins)
}

func mergeWithParent(dat, parent map[string]interface{}) map[string]interface{} {
	for i, v := range dat {
		parent[i] = v
	}
	return parent
}

// Parent can be content or draft.
func ParentIsContent(draft map[string]interface{}) bool {
	if parent_content_i, has_parent_cont := draft["parent_content"]; has_parent_cont && parent_content_i != nil {
		return true
	}
	return false
}

func ParentIsDraft(draft map[string]interface{}) bool {
	if parent_draft_i, has_parent_draft := draft["parent_draft"]; has_parent_draft && parent_draft_i != nil {
		return true
	}
	return false
}

// It's possible that it has no parent at all, then it is a fresh new draft, first version.
func HasNoParent(draft map[string]interface{}) bool {
	if !ParentIsContent(draft) && !ParentIsDraft(draft) {
		return true
	}
	return false
}

// draft_typ example: blog_draft
func BuildDraft(db *mgo.Database, draft_typ, draft_id string) (map[string]interface{}, error) {
	if !strings.HasSuffix(draft_typ, "_draft") { return nil, fmt.Errorf("Draft type does not end with \"_draft\".") }
	d_id := bson.ObjectIdHex(basic.StripId(draft_id))
	q := m{"_id": d_id}
	var v interface{}
	err := db.C("contents").Find(q).One(&v)
	if err != nil { return nil, err }
	if v == nil { return nil, fmt.Errorf("Can't find draft.") }
	draft := basic.Convert(v).(map[string]interface{})
	typ_i, has_typ := draft["type"]
	if !has_typ { return nil, fmt.Errorf("Draft has no type.") }
	typ, is_str := typ_i.(string)
	if !is_str { return nil, fmt.Errorf("Draft type is not a string.") }
	if typ != draft_typ { return nil, fmt.Errorf("Draft type is not the expected one.") }
	var parent_id bson.ObjectId
	if ParentIsContent(draft) {
		parent_id = draft["parent_content"].(bson.ObjectId)
	} else if ParentIsDraft(draft) {
		parent_id = draft["parent_draft"].(bson.ObjectId)
	}
	var par interface{}
	q = m{"_id": parent_id}
	err = db.C("contents").Find(q).One(&par)
	parent := map[string]interface{}{}
	if err == nil || par != nil {
		parent = basic.Convert(par).(map[string]interface{})
	}
	dat := draft["data"].(map[string]interface{})
	return mergeWithParent(dat, parent), nil
}
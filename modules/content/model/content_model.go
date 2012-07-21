package content_model

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/extract"
	"github.com/opesun/jsonp"
	"github.com/opesun/hypecms/model/basic"
	"fmt"
)


func commentRequiredLevel(content_options map[string]interface{}) int {
	var req_lev int
	if lev, has_lev := content_options["comment_level"]; has_lev {
		req_lev = int(lev.(float64))
	} else {
		req_lev = 100
	}
	return req_lev
}

// TODO: it's op independent now.
func AllowsComment(db *mgo.Database, inp map[string][]string, content_options map[string]interface{}, user_id bson.ObjectId, user_level int) error {
	rule := map[string]interface{}{
		"content_id": 	1,
		"comment_id":	1,
	}
	dat, ex_err := extract.New(rule).Extract(inp)
	if ex_err != nil { return ex_err }
	var inserting bool
	if len(dat) == 1 {
		inserting = true
	}
	req_lev := commentRequiredLevel(content_options)
	if req_lev > user_level {
		return fmt.Errorf("You have no rights to comment.")
	}
	// Even if he has the required level, and he is below level 200 (not a moderator), he can't modify other people's comment, only his owns.
	// So we query here the comment and check who is the owner of it.
	if user_level < 200 && !inserting {
		if len(dat) < 2 {
			return fmt.Errorf("Missing fields ", basic.CalcMiss(rule, dat))
		}
		auth, find_err := findCommentAuthor(db, basic.StripId(dat["content_id"].(string)), basic.StripId(dat["comment_id"].(string)))
		if find_err != nil {
			return find_err
		}
		if auth.Hex() != user_id.Hex() {
			return fmt.Errorf("You are not the rightous owner of the comment.")
		}
	}
	return nil
}

// Precedence: type && op, type, op, all
func requiredLevel(content_options map[string]interface{}, typ, op string) int {
	type_op_lev, has := jsonp.Get(content_options, "types." + typ + "." + op + "_level")
	if has {
		return int(type_op_lev.(float64))
	}
	type_lev, has := jsonp.Get(content_options, "types." + typ + ".level")
	if has {
		return int(type_lev.(float64))
	}
	lev, has := content_options["level"]
	if has {
		return int(lev.(float64))
	}
	return 300
}

// op is insert/update/delete
func AllowsContent(db *mgo.Database, inp map[string][]string, content_options map[string]interface{}, user_id bson.ObjectId, user_level int, op string) error {
	rule := map[string]interface{}{
		"type": "must",
		"id": "must",
	}
	dat, ex_err := extract.New(rule).Extract(inp)
	if ex_err != nil { return ex_err }
	var inserting bool
	if len(dat["id"].(string)) == 0 {
		inserting = true
	}
	req_lev := requiredLevel(content_options, dat["type"].(string), op)
	if req_lev > user_level {
		return fmt.Errorf("You have no rights to manage contents.")
	}
	if user_level < 200 && !inserting {
		auth, find_err := findContentAuthor(db, basic.StripId(dat["_id"].(string)))
		if find_err != nil { return find_err }
		if auth.Hex() != user_id.Hex() {
			return fmt.Errorf("You are not the rightous owner of the content.")
		}
	}
	return nil
}

// returns nil if not found
func find(db *mgo.Database, content_id string) map[string]interface{} {
	if len(content_id) != 24 { return nil }
	q := bson.M{
		"_id": bson.ObjectIdHex(content_id),
	}
	var v interface{}
	err := db.C("contents").Find(q).One(&v)
	if err != nil { return nil }
	if v == nil { return nil }
	return basic.Convert(v).(map[string]interface{})
}

func findContentAuthor(db *mgo.Database, content_id string) (bson.ObjectId, error) {
	cont := find(db, content_id)
	if cont == nil {
		return "", fmt.Errorf("Content not found.")
	}
	auth, has := cont["created_by"]
	if !has {
		return "", fmt.Errorf("Content has no author.")
	}
	return auth.(bson.ObjectId), nil
}

func Insert(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId) error {
	id, hasid := dat["id"]
	if hasid && len(id[0]) > 0 {
		return fmt.Errorf("Can't insert an object wich already has an id.")
	}
	typ, hastype := dat["type"]
	if !hastype {
		return fmt.Errorf("No type when inserting content.")
	}
	ins_dat, extr_err := extract.New(rule).Extract(dat)
	if extr_err != nil {
		return extr_err
	}
	basic.DateAndAuthor(rule, ins_dat, user_id)
	ins_dat["type"] = typ[0]
	_, has_tags := ins_dat[Tag_fieldname_displayed]
	if has_tags {
		addTags(db, ins_dat, "", "insert")
	}
	return basic.Inud(db, ev, ins_dat, "contents", "insert", "")
}

func Update(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId) error {
	id, hasid := dat["id"]
	if !hasid {
		return fmt.Errorf("No id when updating content.")
	}
	typ, hastype := dat["type"]
	if !hastype {
		return fmt.Errorf("No type when updating content.")
	}
	upd_dat, extr_err := extract.New(rule).Extract(dat)
	if extr_err != nil {
		return extr_err
	}
	basic.DateAndAuthor(rule, upd_dat, user_id)
	upd_dat["type"] = typ[0]
	_, has_tags := upd_dat[Tag_fieldname_displayed]
	if has_tags {
		addTags(db, upd_dat, id[0], "update")
	}
	return basic.Inud(db, ev, upd_dat, "contents", "update", id[0])
}

func Delete(db *mgo.Database, ev ifaces.Event, id []string, user_id bson.ObjectId) []error {
	var errs []error
	for _, v := range id {
		errs = append(errs, basic.Inud(db, ev, nil, "contents", "delete", v))
	}
	return errs
}

func InsertComment(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, inp map[string][]string, user_id bson.ObjectId) error {
	dat, err := extract.New(rule).Extract(inp)
	if err != nil {
		return err
	}
	basic.DateAndAuthor(rule, dat, user_id)
	ids, err := basic.ExtractIds(inp, []string{"content_id"})
	if err != nil {
		return err
	}
	dat["comment_id"] = bson.NewObjectId()
	q := bson.M{ "_id": bson.ObjectIdHex(ids[0])}
	upd := bson.M{
		"$push": bson.M{
			"comments": dat,
		},
	}
	return db.C("contents").Update(q, upd)
}

// Inp will contain content and comment ID too, as in Update.
func UpdateComment(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, inp map[string][]string, user_id bson.ObjectId) error {
	dat, err := extract.New(rule).Extract(inp)
	if err != nil {
		return err
	}
	basic.DateAndAuthor(rule, dat, user_id)
	ids, err := basic.ExtractIds(inp, []string{"content_id", "comment_id"})
	if err != nil {
		return err
	}
	q := bson.M{
		"_id": bson.ObjectIdHex(ids[0]),
		"comments.comment_id": bson.ObjectIdHex(ids[1]),
	}
	upd := bson.M{
		"$set": bson.M{
			"comments.$": dat,
		},
	}
	return db.C("contents").Update(q, upd)
}

func DeleteComment(db *mgo.Database, ev ifaces.Event, inp map[string][]string, user_id bson.ObjectId) error {
	ids, err := basic.ExtractIds(inp, []string{"content_id", "comment_id"})
	if err != nil {
		return err
	}
	q := bson.M{
		"_id": bson.ObjectIdHex(ids[0]),
		"comments.comment_id": bson.ObjectIdHex(ids[1]),
	}
	upd := bson.M{
		"$pull": bson.M{
			"comments": bson.M{
				"comment_id": bson.ObjectIdHex(ids[1]),
			},
		},
	}
	return db.C("contents").Update(q, upd)
}

// Called from Front hook.
// Find slug value by given key.
func FindContent(db *mgo.Database, keys []string, val string) (map[string]interface{}, bool) {
	query := bson.M{}
	if len(keys) == 0 {
		return nil, false
	} else if len(keys) == 1 {
		if keys[0] == "_id" && len(val) == 24 {			// TODO: check for validity of id.
			query[keys[0]] = bson.ObjectIdHex(val)
		} else {
			query[keys[0]] = val
		}
	} else {
		or := []map[string]interface{}{}
		for _, v := range keys {
			if v == "_id" && len(v) == 24 {				// TODO: check fir validity of id.
				or = append(or, map[string]interface{}{v: bson.ObjectIdHex(val)})
			} else {
				or = append(or, map[string]interface{}{v: val})
			}
		}
		query["$or"] = or
	}
	var v interface{}
	db.C("contents").Find(query).One(&v)
	if v == nil {
		return nil, false
	}
	return basic.Convert(v).(map[string]interface{}), true
}

func findComment(db *mgo.Database, content_id, comment_id string) (map[string]interface{}, error) {
	var v interface{}
	q := bson.M{
		"_id": bson.ObjectIdHex(content_id),
		//"comments.comment_id": bson.ObjectIdHex(comment_id),	
	}
	find_err := db.C("contents").Find(q).One(&v)
	if find_err != nil { return nil, find_err }
	if v == nil {
		return nil, fmt.Errorf("Can't find comment with content id %v, and comment id %v", content_id, comment_id)
	}
	v = basic.Convert(v)
	comments_i, has := v.(map[string]interface{})["comments"]
	if !has {
		return nil, fmt.Errorf("No comments in given content.")
	}
	comments, ok := comments_i.([]interface{})
	if !ok {
		return nil, fmt.Errorf("comments member is not a slice in content %v", content_id)
	}
	// TODO: there must be a better way.
	for _, v_i := range comments {
		v, is_map := v_i.(map[string]interface{})
		if !is_map { continue }
		if val_i, has := v["comment_id"]; has {
			if val_id, ok := val_i.(bson.ObjectId); ok {
				if val_id.Hex() == comment_id {
					return v, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("Comment not found.")
}

func findCommentAuthor(db *mgo.Database, content_id, comment_id string) (bson.ObjectId, error) {
	comment, err := findComment(db, content_id, comment_id)
	if err != nil { return "", err }
	author, has := comment["created_by"]
	if !has {
		return "", fmt.Errorf("Given content has no author.")
	}
	return author.(bson.ObjectId), nil
}
func Install(db *mgo.Database, id bson.ObjectId) error {
	content_options := m{
		"types": m {
			"blog": m{
				"rules" : m{
					"title": 1, "slug":1, "content": 1, Tag_fieldname_displayed : 1,
				},
			},
		},
	}
	q := m{"_id": id}
	upd := m{
		"$addToSet": m{
			"Hooks.Front": "content",
		},
		"$set": m{
			"Modules.content": content_options,
		},
	}
	return db.C("options").Update(q, upd)
}
func Uninstall(db *mgo.Database, id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$pull": m{
			"Hooks.Front": "content",
		},
		"$unset": m{
			"Modules.content": 1,
		},
	}
	return db.C("options").Update(q, upd)
}
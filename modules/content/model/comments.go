package content_model

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/basic"
	"fmt"
	"time"
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

// To be able to list all comments chronologically we insert it to a virtual collection named "comments", where there will be only a link.
func insertToVirtual(db *mgo.Database, content_id, comment_id, author bson.ObjectId) error {
	comment_link := map[string]interface{}{
		"_contents_parent": content_id,
		"comment_id":		comment_id,
		"_users_author":	author,
		"created":			time.Now().Unix(),
	}
	return db.C("comments").Insert(comment_link)
}

func InsertComment(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, inp map[string][]string, user_id bson.ObjectId) error {
	dat, err := extract.New(rule).Extract(inp)
	if err != nil {
		return err
	}
	basic.DateAndAuthor(rule, dat, user_id, false)
	ids, err := basic.ExtractIds(inp, []string{"content_id"})
	if err != nil {
		return err
	}
	content_id := bson.ObjectIdHex(ids[0])
	comment_id := bson.NewObjectId()
	dat["comment_id"] = comment_id
	q := bson.M{ "_id": content_id}
	upd := bson.M{
		"$push": bson.M{
			"comments": dat,
		},
	}
	err = db.C("contents").Update(q, upd)
	if err == nil {
		insertToVirtual(db, content_id, comment_id, user_id) 
	}
	return err
}

// Inp will contain content and comment ID too, as in Update.
func UpdateComment(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, inp map[string][]string, user_id bson.ObjectId) error {
	dat, err := extract.New(rule).Extract(inp)
	if err != nil {
		return err
	}
	basic.DateAndAuthor(rule, dat, user_id, true)
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
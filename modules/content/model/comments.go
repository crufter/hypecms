package content_model

import (
	"fmt"
	"github.com/opesun/extract"
	ifaces "github.com/opesun/hypecms/frame/interfaces"
	"github.com/opesun/hypecms/frame/misc/basic"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

// Only called when doing update or delete.
// At inserting, user.OkayToDoAction is sufficient.
// This needs additional action: you can only update or delete a given comment only if it's yours (under a certain level).
func CanModifyComment(db *mgo.Database, inp map[string][]string, correction_level int, user_id bson.ObjectId, user_level int) error {
	rule := map[string]interface{}{
		"content_id": "must",
		"comment_id": "must",
		"type":       "must",
	}
	dat, err := extract.New(rule).Extract(inp)
	if err != nil {
		return err
	}
	// Even if he has the required level, and he is below correction_level, he can't modify other people's comment, only his owns.
	// So we query here the comment and check who is the owner of it.
	if user_level < correction_level {
		content_id_str := basic.StripId(dat["content_id"].(string))
		comment_id_str := basic.StripId(dat["comment_id"].(string))
		auth, err := findCommentAuthor(db, content_id_str, comment_id_str)
		if err != nil {
			return err
		}
		if auth.Hex() != user_id.Hex() {
			return fmt.Errorf("You are not the rightous owner of the comment.")
		}
	}
	return nil
}

// To be able to list all comments chronologically we insert it to a virtual collection named "comments", where there will be only a link.
// "_id" equals to "comment_id" in the content comment array.
func insertToVirtual(db *mgo.Database, content_id, comment_id, author bson.ObjectId, typ string, in_moderation bool) error {
	comment_link := map[string]interface{}{
		"_contents_parent": content_id,
		"_users_author":    author,
		"created":          time.Now().Unix(),
		"content_type":		typ,
		"in_moderation":	in_moderation,
	}
	if in_moderation {
		comment_link["_comments_moderation"] = comment_id
	} else {
		comment_link["comment_id"] = comment_id
	}
	return db.C("comments").Insert(comment_link)
}

// Places a comment into its final place - the comment array field of a given content.
func insertToFinal(db *mgo.Database, comment map[string]interface{}, comment_id, content_id bson.ObjectId) error {
	comment["comment_id"] = comment_id
	q := bson.M{"_id": content_id}
	upd := bson.M{
		"$inc": bson.M{
			"comment_count": 1,
		},
		"$push": bson.M{
			"comments": comment,
		},
	}
	return db.C("contents").Update(q, upd)
}

// MoveToFinal with extract.
func MoveToFinalWE(db *mgo.Database, inp map[string][]string) error {
	r := map[string]interface{}{
		"comment_id": "must",
	}
	dat, err := extract.New(r).Extract(inp)
	if err != nil {
		return err
	}
	comment_id := basic.ToIdWithCare(dat["comment_id"])
	return MoveToFinal(db, comment_id)
}

// Moves a comment to its final destination (into the valid comments) from moderation queue.
func MoveToFinal(db *mgo.Database, comment_id bson.ObjectId) error {
	var comm interface{}
	q := m{"_id": comment_id}
	err := db.C("comments_moderation").Find(q).One(&comm)
	if err != nil {
		return err
	}
	comment := basic.Convert(comm).(map[string]interface{})
	comment["comment_id"] = comment["_id"]
	delete(comment, "comment_id")
	content_id := comment["_contents_parent"].(bson.ObjectId)
	q2 := m{"_id": content_id}
	upd2 := m{
		"$inc": m{
			"comment_count": 1,
		},
		"$push": m{
			"comments": comment,
		},
	}
	err = db.C("contents").Update(q2, upd2)
	if err != nil {
		return err
	}
	upd := m{
		"$set": m{
			"in_moderation": false,
		},
	}
	err = db.C("comments").Update(q, upd)
	if err != nil {
		return err
	}
	return db.C("comments_moderation").Remove(q)
}

// Maybe we should just delete the comment in this case?
// Think about it and implement it later.
func MoveToModeration(db *mgo.Database, content_id, comment_id bson.ObjectId) error {
	return fmt.Errorf("Not implemented yet.")
}

// Puts comment coming from UI into moderation queue.
func insertModeration(db *mgo.Database, comment map[string]interface{}, comment_id, content_id bson.ObjectId, typ string) error {
	comment["_id"] = comment_id
	comment["_contents_parent"] = content_id
	comment["content_type"] = typ
	return db.C("comments_moderation").Insert(comment)
}

// Apart from rule, there is one mandatory field which must come from the UI: "content_id"
// moderate_first should be read as "moderate first if it is a valid, spam protection passed comment"
// Spam protection happens outside of this anyway.
func InsertComment(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, inp map[string][]string, user_id bson.ObjectId, typ string, moderate_first bool) error {
	dat, err := extract.New(rule).Extract(inp)
	if err != nil {
		return err
	}
	dat["type"] = typ
	basic.DateAndAuthor(rule, dat, user_id, false)
	ids, err := basic.ExtractIds(inp, []string{"content_id"})
	if err != nil {
		return err
	}
	content_id := bson.ObjectIdHex(ids[0])
	comment_id := bson.NewObjectId()
	if moderate_first {
		err = insertModeration(db, dat, comment_id, content_id, typ)
	} else {
		err = insertToFinal(db, dat, comment_id, content_id)
	}
	// This will be made optional, a facebook style app does not need it, only a bloglike site.
	if err == nil {
		err = insertToVirtual(db, content_id, comment_id, user_id, typ, moderate_first)
	}
	return err
}

// Apart from rule, there are two mandatory field which must come from the UI: "content_id" and "comment_id"
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
	comment_id := bson.ObjectIdHex(ids[1])
	q := bson.M{
		"_id": bson.ObjectIdHex(ids[0]),
		"comments.comment_id": comment_id,
	}
	upd := bson.M{
		"$set": bson.M{
			"comments.$": dat,
		},
	}
	err = db.C("contents").Update(q, upd)
	if err != nil {
		return err
	}
	return db.C("comments").Remove(m{"_id": comment_id})
}

// Two mandatory fields must come from UI: "content_id" and "comment_id"
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
		"$inc": bson.M{
			"comment_count": -1,
		},
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
	if find_err != nil {
		return nil, find_err
	}
	if v == nil {
		return nil, fmt.Errorf("Can't find content with id %v.", content_id)
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
		if !is_map {
			continue
		}
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
	if err != nil {
		return "", err
	}
	author, has := comment["created_by"]
	if !has {
		return "", fmt.Errorf("Given content has no author.")
	}
	return author.(bson.ObjectId), nil
}

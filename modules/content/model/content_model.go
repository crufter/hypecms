package content_model

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/basic"
	"fmt"
)

func Find(db *mgo.Database, content_id string) {
	
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
	upd_dat["typ"] = typ[0]
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
	return db.C("contents").Update(bson.M{"_id": bson.ObjectIdHex(ids[0])}, bson.M{"$push": bson.M{"comments": dat}})
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
	return db.C("contents").Update(bson.M{"_id": bson.ObjectIdHex(ids[0]), "comments.comment_id": bson.ObjectIdHex(ids[1])}, bson.M{"$set": bson.M{"comments.$": dat}})
}

func DeleteComment(db *mgo.Database, ev ifaces.Event, inp map[string][]string, user_id bson.ObjectId) error {
	ids, err := basic.ExtractIds(inp, []string{"content_id", "comment_id"})
	if err != nil {
		return err
	}
	return db.C("contents").Update(bson.M{"_id": bson.ObjectIdHex(ids[0]), "comments.comment_id": bson.ObjectIdHex(ids[1])}, bson.M{"$pull": bson.M{"comments": bson.M{"comment_id": bson.ObjectIdHex(ids[1])}}})
}

// Find slug value be given key.
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
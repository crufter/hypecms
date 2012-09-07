package content_model

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"testing"
)

type Event struct {
}

func (e Event) Param(params ...interface{}) {
}

func (e Event) TriggerAll(eventnames ...string) {
}

func (e Event) Trigger(eventname string, params ...interface{}) {
}

func TestTags(t *testing.T) {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		t.Fatal("Cant connect to db.")
	}
	db := session.DB("tags_test")
	ev := &Event{}
	rule := map[string]interface{}{
		tag_fieldname_displayed: 1,
		"type":                  1,
	}
	dat := map[string][]string{
		tag_fieldname_displayed: {"lol, lal, lol, lal"},
		"type":                  {"blog"},
	}
	uid := bson.NewObjectId()
	// id := bson.NewObjectId()
	// dat["_id"] = id
	err = Insert(db, ev, rule, dat, uid)
	if err != nil {
		t.Fatal(err.Error())
	}
	var i []interface{}
	db.C(collection_name).Find(nil).All(&i)
	if len(i) != 2 {
		t.Fatal("Bad number of tags: ", len(i))
	}
	for _, v := range i {
		val := v.(bson.M)
		counter := val[count_fieldname].(int)
		if counter != 1 {
			t.Fatal("Bad tag count: ", counter, " instead of 1.")
		}
	}
	err = db.DropDatabase()
	if err != nil {
		t.Fatal("Cant drop database.")
	}
}

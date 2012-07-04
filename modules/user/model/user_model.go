package user_model

import(
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"github.com/opesun/hypecms/model"
	"fmt"
)

func FindUser(db *mgo.Database, id string) (map[string]interface{}, bool) {
	var v interface{}
	db.C("users").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&v)
	if v != nil {
		val, _ := model.Convert(v).(map[string]interface{})
		delete(val, "password")
		return val, true
	}
	return nil, false
}

func FindLogin(db *mgo.Database, name, encoded_pass string) (string, bool) {
	var v interface{}
	db.C("users").Find(bson.M{"name": name, "password": encoded_pass}).One(&v)
	if v != nil {
		return v.(bson.M)["_id"].(bson.ObjectId).Hex(), true
	}
	return "", false
}

func BuildUser(db *mgo.Database, ev model.Event, user_id string) map[string]interface{} {
	var user map[string]interface{}
	var ok bool
	if len(user_id) > 0 {
		user, ok = FindUser(db, user_id)
	}
	if !ok {
		user = make(map[string]interface{})
		user["level"] = 0
	}
	ev.Trigger("user.build", user)
	return user
}

// We should call the extract module here, also no name && pass but rather a map[string]interface{} containing all the things.
func Register(db *mgo.Database, ev model.Event, name, pass string) error {
	u := bson.M{"name": name, "password": pass}
	ev.Trigger("user.register", u)
	err := db.C("users").Insert(u)
	if err != nil {
		return fmt.Errorf("name is not unique")
	}
	return nil
}


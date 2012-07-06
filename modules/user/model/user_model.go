package user_model

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/basic"
	ifaces "github.com/opesun/hypecms/interfaces"
	"fmt"
)

func FindUser(db *mgo.Database, id string) (map[string]interface{}, error) {
	v:= basic.Find(db, "users", id)
	if v != nil {
		delete(v, "password")
		return v, nil
	}
	return nil, fmt.Errorf("Can't find user with id " + id)
}

func NamePass(db *mgo.Database, name, encoded_pass string) (map[string]interface{}, error) {
	var v interface{}
	db.C("users").Find(bson.M{"name": name, "password": encoded_pass}).One(&v)
	if v != nil {
		return basic.Convert(v).(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("Can't find user/password combo.")
}

func Login(db *mgo.Database, inp map[string][]string) (map[string]interface{}, string, error) {
	rule := map[string]interface{}{
		"name": map[string]interface{}{
			"must": 1,
			"type":	"string",
		},
		"password": map[string]interface{}{
			"must": 1,
			"type": "string",
		},
	}
	d, err := extract.New(rule).Extract(inp)
	if err != nil {
		return nil, "", err
	}
	user, err := NamePass(db, d["name"].(string), d["password"].(string))
	if err != nil {
		return nil, "", err
	}
	return user, user["_id"].(bson.ObjectId).Hex(), nil
}

func BuildUser(db *mgo.Database, ev ifaces.Event, user_id string) map[string]interface{} {
	var user map[string]interface{}
	var err error
	if len(user_id) > 0 {
		user, err = FindUser(db, user_id)
	}
	if err != nil {
		user = make(map[string]interface{})
		user["level"] = 0
	}
	ev.Trigger("user.build", user)
	return user
}

// We should call the extract module here, also no name && pass but rather a map[string]interface{} containing all the things.
func Register(db *mgo.Database, ev ifaces.Event, name, pass string) error {
	u := bson.M{"name": name, "password": pass}
	err := db.C("users").Insert(u)
	if err != nil {
		return fmt.Errorf("Name is not unique.")
	}
	ev.Trigger("user.register", u)
	return nil
}
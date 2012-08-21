package user_model

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/basic"
	ifaces "github.com/opesun/hypecms/interfaces"
	"crypto/sha1"
	"fmt"
	"io"
	"strings"
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
		"name": "must",
		"password": "must",
	}
	d, err := extract.New(rule).Extract(inp)
	if err != nil {
		return nil, "", err
	}
	user, err := NamePass(db, d["name"].(string), Encode(d["password"].(string)))
	if err != nil {
		return nil, "", err
	}
	return user, user["_id"].(bson.ObjectId).Hex(), nil
}

func EmptyUser() map[string]interface{} {
	user := make(map[string]interface{})
	user["level"] = 0
	return user
}

func ParseAcceptLanguage(l string) []string {
	ret := []string{}
	sl := strings.Split(l, ",")
	c := map[string]struct{}{}
	for _, v := range sl {
		lang := string(strings.Split(v, ";")[0][0:2])
		_, has := c[lang]
		if !has {
			c[lang] = struct{}{}
			ret = append(ret, lang)
		}
	}
	return ret
}

func BuildUser(db *mgo.Database, ev ifaces.Event, user_id string, http_header map[string][]string) map[string]interface{} {
	var user map[string]interface{}
	var err error
	if len(user_id) > 0 {
		user, err = FindUser(db, user_id)
	}
	if err != nil || user == nil {
		user = EmptyUser()
	}
	_, langs_are_set := user["languages"]
	if !langs_are_set {
		langs, has := http_header["Accept-Language"]
		if has {
			user["languages"] = ParseAcceptLanguage(langs[0])
		} else {
			user["languages"] = []string{"en"}
		}
	}
	ev.Trigger("user.build", user)
	return user
}

func Encode(pass string) string {
	h := sha1.New()
	io.WriteString(h, pass)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// We should call the extract module here, also no name && pass but rather a map[string]interface{} containing all the things.
func Register(db *mgo.Database, ev ifaces.Event, name, pass string) error {
	u := bson.M{"name": name, "password": Encode(pass)}
	err := db.C("users").Insert(u)
	if err != nil {
		return fmt.Errorf("Name is not unique.")
	}
	ev.Trigger("user.register", u)
	return nil
}
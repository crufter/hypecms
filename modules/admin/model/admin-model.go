package admin_model

import (
	"encoding/json"
	"fmt"
	"github.com/opesun/extract"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/hypecms/modules/user/model"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type m map[string]interface{}

// consts for regUser modes.
const (
	first_admin     = 1
	admin_with_name = 2
	generic_user    = 3
)

func SiteHasAdmin(db *mgo.Database) bool {
	var v interface{}
	db.C("users").Find(m{"level": m{"$gt": 299}}).One(&v)
	return v != nil
}

// mode: 1	first admin of site, named "admin"
// mode: 2	admin with a name
// mode: 3	a generic user
// Database must have a unique index on slugs to avoid user slug duplications.
func regUser(db *mgo.Database, post map[string][]string, mode int) error {
	subr := map[string]interface{}{
		"must": 1,
		"type": "string",
		"min":  1,
	}
	rule := map[string]interface{}{
		"password":       subr,
		"password_again": subr,
	}
	if mode != first_admin {
		rule["name"] = subr
	}
	dat, err := extract.New(rule).Extract(post)
	if err != nil {
		return err
	}
	pass := dat["password"].(string)
	pass_again := dat["password_again"].(string)
	if pass != pass_again {
		return fmt.Errorf("Password and password confirmation differs.")
	}
	a := map[string]interface{}{"password": user_model.EncodePass(pass)}
	switch mode { // Redundant in places for better readability.
	case first_admin:
		a["name"] = "admin"
		a["level"] = 300
	case admin_with_name:
		a["name"] = dat["name"]
		a["level"] = 300
	case generic_user:
		a["name"] = dat["name"]
		a["level"] = 100
	}
	err = db.C("users").Insert(a)
	if err != nil {
		return fmt.Errorf("Name is not unique.")
	}
	return nil
}

// Registers a user, without fluff, only name and password.
func RegUser(db *mgo.Database, post map[string][]string) error {
	return regUser(db, post, generic_user)
}

// Registers the first administrator of the site, with fixed name "admin".
func RegFirstAdmin(db *mgo.Database, post map[string][]string) error {
	return regUser(db, post, first_admin)
}

func RegAdmin(db *mgo.Database, post map[string][]string) error {
	return regUser(db, post, admin_with_name)
}

// opt structure:
// Modules.modulename
func InstallB(db *mgo.Database, ev ifaces.Event, opt map[string]interface{}, modn, mode string) (bson.ObjectId, error) {
	var object_id bson.ObjectId
	if _, already := jsonp.Get(opt, "Modules."+modn); mode == "install" && already {
		return object_id, fmt.Errorf("Module " + modn + " is already installed.")
	} else if mode == "uninstall" && !already {
		return object_id, fmt.Errorf("Module " + modn + " is not installed.")
	}
	object_id = basic.CreateOptCopy(db)
	return object_id, nil
}

func SaveConfig(db *mgo.Database, ev ifaces.Event, encoded_conf string) error {
	var v interface{}
	json.Unmarshal([]byte(encoded_conf), &v)
	if v != nil {
		m := v.(map[string]interface{})
		// Just in case
		delete(m, "_id")
		object_id := basic.CreateOptCopy(db)
		m["created"] = time.Now().Unix()
		db.C("options").Update(bson.M{"_id": object_id}, m)
	} else {
		return fmt.Errorf("Invalid json.")
	}
	return nil
}

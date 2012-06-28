// Package user implements basic user functionality.
// - Registration, deletion, update, login, logout of users.
// - Building the user itself (if logged in), and putting it to uni.Dat["_user"].
package user

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"net/http"
	"fmt"
)

var Hooks = map[string]func(*context.Uni) error {
	"BuildUser": BuildUser,
	"Back":      Back,
	"Test":      Test,
}

func FindUser(db *mgo.Database, id string) (map[string]interface{}, bool) {
	var v interface{}
	db.C("users").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&v)
	if v != nil {
		val, _ := context.Convert(v).(map[string]interface{})
		delete(val, "password")
		return val, true
	}
	return nil, false
}

func BuildUser(uni *context.Uni) error {
	c, err := uni.Req.Cookie("user")
	var user map[string]interface{}
	var ok bool
	if err == nil && len(c.Value) > 0 {
		user, ok = FindUser(uni.Db, c.Value)
	}
	if ok {
		uni.Dat["_user"] = user
	} else {
		m := make(map[string]interface{})
		m["level"] = 0
		uni.Dat["_user"] = m
	}
	return nil
}

func FindLogin(db *mgo.Database, name, encoded_pass string) (string, bool) {
	var v interface{}
	db.C("users").Find(bson.M{"name": name, "password": encoded_pass}).One(&v)
	if v != nil {
		return v.(bson.M)["_id"].(bson.ObjectId).Hex(), true
	}
	return "", false
}

func Register(uni *context.Uni) error {
	post := uni.Req.Form
	name, name_ok := post["name"]
	pass, pass_ok := post["password"]
	if name_ok && pass_ok && len(name) > 0 && len(pass) > 0 {
		u := bson.M{"name": name[0], "password": pass[0]}
		// Ide jön, hogy kiszedjük opciókból hogy miket pakoljunk még bele etc...
		err := uni.Db.C("users").Insert(u)
		if err != nil {
			return fmt.Errorf("name is not unique")
		}
	} else {
		return fmt.Errorf("no name given")
	}
	return nil
}

func Login(uni *context.Uni) error {
	// There could be a check here to not log in somebody who is already logged in.
	name, name_ok := uni.Req.Form["name"]
	pass, pass_ok := uni.Req.Form["password"]
	succ := false
	reason := make([]string, 0)
	if name_ok && pass_ok && len(name) == 1 && len(pass) == 1 {
		name_str := name[0]
		pass_str := pass[0]
		if id, ok := FindLogin(uni.Db, name_str, pass_str); ok {
			c := &http.Cookie{Name: "user", Value: id, MaxAge: 3600000, Path: "/"}
			http.SetCookie(uni.W, c)
			succ = true
		} else {
			reason = append(reason, "cant find user/password combo")
		}
	} else {
		if !name_ok {
			reason = append(reason, "no name given")
		} else if len(name) != 1 {
			reason = append(reason, "improper name")
		}
		if !pass_ok {
			reason = append(reason, "no pass given")
		} else if len(pass) != 1 {
			reason = append(reason, "improper pass")
		}
	}
	if !succ {
		return fmt.Errorf(reason[0])	// Ugly hack now, because main.handleBacks expects a string, not a []string.
	}
	return nil
}

func Logout(uni *context.Uni) error {
	c := &http.Cookie{Name: "user", Value: "", Path: "/"}
	http.SetCookie(uni.W, c)
	return nil
}

func TestRaw(opt map[string]interface{}) map[string]interface{} {
	msg := make(map[string]interface{})
	// _, has := jsonp.Get(opt, "BuildUser")
	// msg["BuildUser"] = has
	has := jsonp.HasVal(opt, "Hooks.Back", "user")
	msg["Back"] = has
	return msg
}

func Test(uni *context.Uni) error {
	uni.Dat["_cont"] = TestRaw(uni.Opt)
	return nil
}

func Back(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
	had_action := true
	switch action {
	case "login":
		Login(uni)
	case "logout":
	case "register":
	default:
		had_action = false
	}
	if !had_action {
		uni.Put("Can't find action named \"" + action + "\" in user module.")
	}
	return nil
}

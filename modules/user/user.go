// Package user implements basic user functionality.
// - Registration, deletion, update, login, logout of users.
// - Building the user itself (if logged in), and putting it to uni.Dat["_user"].
package user

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/user/model"
	"github.com/opesun/jsonp"
	"net/http"
	"fmt"
)

var Hooks = map[string]func(*context.Uni) error {
	"BuildUser": BuildUser,
	"Back":      Back,
	"Test":      Test,
}

func BuildUser(uni *context.Uni) error {
	c, _ := uni.Req.Cookie("user")
	uni.Dat["_user"] = user_model.BuildUser(uni.Db, uni.Ev, c.Value)
	return nil
}

func Register(uni *context.Uni) error {
	post := uni.Req.Form
	name, name_ok := post["name"]
	pass, pass_ok := post["password"]
	if name_ok && pass_ok && len(name) > 0 && len(pass) > 0 {
		return user_model.Register(uni.Db, uni.Ev, name[0], pass[0])
	} else {
		return fmt.Errorf("No name or pass given.")
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
		if id, ok := user_model.FindLogin(uni.Db, name_str, pass_str); ok {
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
	var err error
	switch action {
	case "login":
		err = Login(uni)
	case "logout":
	case "register":
	default:
		err = fmt.Errorf("Unkown action at user module.")
	}
	return err
}

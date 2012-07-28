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
	c, err := uni.Req.Cookie("user")
	if err == nil {
		uni.Dat["_user"] = user_model.BuildUser(uni.Db, uni.Ev, c.Value)
	} else {
		uni.Dat["_user"] = user_model.EmptyUser()
	}
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
	inp := map[string][]string(uni.Req.Form)
	if _, id, err := user_model.Login(uni.Db, inp); err == nil {
		c := &http.Cookie{Name: "user", Value: id, MaxAge: 3600000, Path: "/"}
		http.SetCookie(uni.W, c)
	} else {
		return err
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

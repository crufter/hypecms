// Package user implements basic user functionality.
// - Registration, deletion, update, login, logout of users.
// - Building the user itself (if logged in), and putting it to uni.Dat["_user"].
package user

import (
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/user/model"
	"github.com/opesun/jsonp"
	"net/http"
)

var Hooks = map[string]interface{}{
	"BuildUser": BuildUser,
	"Back":      Back,
	"Test":      Test,
}

// Recover from wrong ObjectId like panics. Unset the cookie.
func unsetCookie(w http.ResponseWriter, dat map[string]interface{}, err *error) {
	r := recover()
	if r == nil {
		return
	}
	*err = nil // Just to be sure.
	c := &http.Cookie{Name: "user", Value: "", MaxAge: 3600000, Path: "/"}
	http.SetCookie(w, c)
	dat["_user"] = user_model.EmptyUser()
}

// If there were some random database query errors or something we go on with an empty user.
func BuildUser(uni *context.Uni) (err error) {
	defer unsetCookie(uni.W, uni.Dat, &err)
	var user_id_str string
	c, err := uni.Req.Cookie("user")
	if err != nil {
		panic(err)
	}
	user_id_str = c.Value
	block_key := []byte(uni.Secret())
	user_id, err := user_model.DecryptId(user_id_str, block_key)
	if err != nil {
		panic(err)
	}
	user, err := user_model.BuildUser(uni.Db, uni.Ev, user_id, uni.Req.Header)
	if err != nil {
		panic(err)
	}
	uni.Dat["_user"] = user
	return
}

func Register(uni *context.Uni) error {
	inp := uni.Req.Form
	rules, _ := jsonp.GetM(uni.Opt, "Modules.user.rules") // RegisterUser will be fine with nil.
	_, err := user_model.RegisterUser(uni.Db, uni.Ev, rules, inp)
	return err
}

func Login(uni *context.Uni) error {
	// Maybe there could be a check here to not log in somebody who is already logged in.
	inp := uni.Req.Form
	if _, id, err := user_model.FindLogin(uni.Db, inp); err == nil {
		block_key := []byte(uni.Secret())
		return user_model.Login(uni.W, id, block_key)
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

func Back(uni *context.Uni, action string) error {
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

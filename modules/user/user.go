// Package user implements basic user functionality.
// - Registration, deletion, update, login, logout of users.
// - Building the user itself (if logged in), and putting it to uni.Dat["_user"].
package user

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/user/model"
	"github.com/opesun/jsonp"
	"net/http"
)

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

type H struct {
	uni *context.Uni
}

func Hooks(uni *context.Uni) *H {
	return &H{uni}
}

type A struct {
	uni *context.Uni
}

func Actions(uni *context.Uni) *A {
	return &A{uni}
}

// If there were some random database query errors or something we go on with an empty user.
func (h *H) BuildUser() (err error) {
	uni := h.uni
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

func (a *A) Register() error {
	inp := a.uni.Req.Form
	rules, _ := jsonp.GetM(a.uni.Opt, "Modules.user.rules") // RegisterUser will be fine with nil.
	_, err := user_model.RegisterUser(a.uni.Db, a.uni.Ev, rules, inp)
	return err
}

func (a *A) Login() error {
	// Maybe there could be a check here to not log in somebody who is already logged in.
	inp := a.uni.Req.Form
	if _, id, err := user_model.FindLogin(a.uni.Db, inp); err == nil {
		block_key := []byte(a.uni.Secret())
		return user_model.Login(a.uni.W, id, block_key)
	} else {
		return err
	}
	return nil
}

func (a *A) Logout() error {
	c := &http.Cookie{Name: "user", Value: "", Path: "/"}
	http.SetCookie(a.uni.W, c)
	return nil
}

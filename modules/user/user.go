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

// Helper function to hotregister a guest user, log him in and build his user data into uni.Dat["_user"].
func RegLoginBuild(uni *context.Uni) error {
	db := uni.Db
	ev := uni.Ev
	guest_rules, _ := jsonp.GetM(uni.Opt, "Modules.user.guest_rules") // RegksterGuest will do fine with nil.
	inp := uni.Req.Form
	http_header := uni.Req.Header
	dat := uni.Dat
	w := uni.W
	block_key := []byte(uni.Secret())
	guest_id, err := user_model.RegisterGuest(db, ev, guest_rules, inp)
	if err != nil {
		return err
	}
	_, _, err = user_model.FindLogin(db, inp)
	if err != nil {
		return err
	}
	err = user_model.Login(w, guest_id, block_key)
	if err != nil {
		return err
	}
	user, err := user_model.BuildUser(db, ev, guest_id, http_header)
	if err != nil {
		return err
	}
	dat["_user"] = user
	return nil
}

func Honeypot(uni *context.Uni, options map[string]interface{}) error {
	return nil
}

func Hashcash(uni *context.Uni, options map[string]interface{}) error {
	return nil
}

// It must be called like this:
// user.PuzzleSolved(uni, "content.types.blog.comment_insert")
// Then, Sprintf("Modules.%v_puzzles", puzzle_path) will be queried and executed.
//
// This will definitely need some refinement, but we try to keep it simple for now, even if it will not fit all use cases.
func PuzzleSolved(uni *context.Uni, path string) error {
	locate := fmt.Sprintf("Modules.%v_puzzles", path)
	puzzle_group_i, ok := jsonp.GetS(uni.Opt, locate)
	if !ok {
		return fmt.Errorf("Can't find puzzle names. Returning, because your system is unsecure.") // We return an error here just to be sure.
	}
	can_fail := 0 // How manny puzzle one can fail before returning an error.
	puzzle_group, can_fail := user_model.PuzzleGroup(puzzle_group_i)
	failed := 0
	failed_puzzles := []string{}
	for _, v := range puzzle_group {
		if failed > can_fail {
			return fmt.Errorf("Failed more than %v puzzles. Failed: %v.", can_fail, failed_puzzles)
		}
		puzzle_locate := fmt.Sprintf("Modules.user.puzzles.%v", v)
		puzzle_opt, ok := jsonp.GetM(uni.Opt, puzzle_locate)
		if !ok {
			return fmt.Errorf("Cant find puzzle named %v.", v)
		}
		var err error
		switch v {
		case "hascash":
			err = Hashcash(uni, puzzle_opt)
		case "honeypot":
			err = Honeypot(uni, puzzle_opt)
		}
		if err != nil {
			failed_puzzles = append(failed_puzzles, v)
			failed++
		}
	}
	return nil
}

func ShowHashcash(uni *context.Uni) (string, error) {
	return "", nil
}

func ShowHoneypot(uni *context.Uni) (string, error) {
	return "", nil
}

func ShowPuzzle(uni *context.Uni, puzzle_path string) (string, error) {
	return "", nil
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

// Thin controller-like layer supporting main.go
// Currently contains mostly authentication-related functions.
//
// Currently user levels are:
// 0: total stranger.
// 1: someone who already done an action, was "registered on the fly", but failed puzzles
// 2: someone who already done an action, was registered on the fly and solved the puzzles successfully
// 100: registered user
// Above this, user levels are not well defined yet:
// 200: moderator-like entity
// 300: admin, full rights.
package user

import(
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/user/model"
	"github.com/opesun/jsonp"
	"github.com/opesun/numcon"
)

const(
	cant_run_back	= "Can't run back hook %v: module %v is not installed."
	not_impl		= "Not implemented yet."
)

func defaultPuzzles(uni *context.Uni) []interface{} {
	def, has := jsonp.GetS(uni.Opt, "user.default_puzzles")
	if !has || len(def) == 0 {
		return []interface{}{"timer"}
	}
	return def
}

// Writes not existing default values to the auth_options.
func authDefaults(uni *context.Uni, auth_o map[string]interface{}) map[string]interface{} {
	if auth_o == nil {
		auth_o = map[string]interface{}{}
	}
	if _, has := auth_o["min_lev"]; !has {
		auth_o["min_lev"] = 300
	}
	if _, has := auth_o["no_puzzles_lev"]; !has {
		auth_o["no_puzzles_lev"] = 2
	}
	if _, has := auth_o["puzzles"]; !has {
		auth_o["puzzles"] = defaultPuzzles(uni)
	}
	if _, has := auth_o["hot_reg"]; !has {
		auth_o["hot_reg"] = 0
	}
	return auth_o
}

// Retrieves the map which drives the given authorization from the option document.
func AuthOpts(uni *context.Uni, mod_name, action_name string) (auth_opts map[string]interface{}, explicit_ignore bool) {
	val, has := jsonp.Get(uni.Opt, fmt.Sprintf("Modules.%v.actions.%v.auth", mod_name, action_name))
	if !has {
		return authDefaults(uni, nil), false
	}
	boolval, isbool := val.(bool)
	if isbool && boolval == false {
		return nil, true
	}
	auth_opts, ok := val.(map[string]interface{})
	if !ok {
		return authDefaults(uni, nil), false
	}
	return authDefaults(uni, auth_opts), false
}

// A very basic framework to provide an easy way to do action based authorization (currently checks user levels and puzzles).
// Hopefully this will solve the common security problem of forgetting to check the user's rights in modules,
// since everything is blacklisted by default (needs admin rights).
//
// Example:
// "Modules.%v.actions.%v.auth" : {
// 		"min_lev": 0,				// Defaults to 300. 0 Means somebody who has a user level >= min_lev can do it.
//		"no_puzzles_lev": 2			// Defaults to 2. Means someone who has a user level >= no_puzzles_lev will not have to solve the spam protection puzzle.
//		"puzzles": ["timer"]		// Defaults to defaultPuzzles(uni).
//		"hot_reg": 2				// More precisely: "reg, login, build".
//									// Defaults to 0. Specifies wether to register, login and build a guest user.
//									// 0 means don't register at all. 1 means register if he solved the puzzles. 2 register even if he failed the puzzles (useful for moderation).
// }
//
// A value of false means proceed as passed. This is useful when the rights to an action can not be determined by only
// from the module and action name. A good example is the content module. An action of "insert", or "comment_insert" can belong
// to different types of content, thus requiring different levels.
// We can solve this problem by assigning "Modules.content.actions.insert.auth" = false
// and calling this function by hand as mod_name = "content.types.blog", action_name = "insert" => "Modules.content.types.blog.actions.insert.auth" (long, I know...).
//
// Better workaround must exists, but currently we go on with this in the content module.
// First error is general error, not meant to be ignored, second is puzzle error, which can be ignored if one wants implement moderation.
func OkayToDoAction(uni *context.Uni, mod_name, action_name string) (error, error) {
	if _, installed := jsonp.Get(uni.Opt, "Modules." + mod_name); !installed {
		return fmt.Errorf(cant_run_back, action_name, mod_name), nil
	}
	auth_options, explicit_ignore := AuthOpts(uni, mod_name, action_name)
	if explicit_ignore {
		return nil, nil
	}
	return AuthAction(uni, auth_options)
}

// Similar to OkayToDoAction but it works directly on the auth_options map.
func AuthAction(uni *context.Uni, auth_options map[string]interface{}) (error, error) {
	err := UserAllowed(uni, auth_options)
	if err != nil {
		return err, nil
	}
	user_level := scut.Ulev(uni.Dat["_user"])
	no_puzzles_lev := numcon.IntP(auth_options["no_puzzles_lev"])
	var hot_reg int
	if val, has := auth_options["hot_reg"]; has {
		hot_reg = numcon.IntP(val)
	}
	var puzzle_err error
	if user_level < no_puzzles_lev {
		puzzle_err = SolvePuzzles(uni, auth_options)
	}
	if user_level == 0 && ((puzzle_err == nil && hot_reg >= 1) || (puzzle_err != nil && hot_reg == 2)) {
		err = RegLoginBuild(uni, puzzle_err == nil)
	}
	return err, puzzle_err
}

func guestRules(uni *context.Uni) map[string]interface{} {
	rules, has := jsonp.GetM(uni.Opt, "user.guest_rules") // RegksterGuest will do fine with nil.
	if has {
		return rules
	}
	return map[string]interface{}{
		"guest_name": map[string]interface{}{
			"must": true,
			"min": 1,
			"max": 50,
		},
		"website": 1,
	}
}

// Helper function to hotregister a guest user, log him in and build his user data into uni.Dat["_user"].
func RegLoginBuild(uni *context.Uni, solved_puzzle bool) error {
	db := uni.Db
	ev := uni.Ev
	guest_rules := guestRules(uni)
	inp := uni.Req.Form
	http_header := uni.Req.Header
	dat := uni.Dat
	w := uni.W
	block_key := []byte(uni.Secret())
	guest_id, err := user_model.RegisterGuest(db, ev, guest_rules, inp, solved_puzzle)
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

// Immediately terminate the run of the action in case the user level is lower than the required level of the given action.
// By default, if not otherwise specified, every action requires a level of 300 (admin rights).
//
// Made public to be able to call separately from PuzzlesSolved.
// This way one can implement moderation.
func UserAllowed(uni *context.Uni, auth_options map[string]interface{}) error {
	minlev := 300
	lev_in_opt := auth_options["min_lev"]
	num, err := numcon.Int(lev_in_opt)
	if err == nil {
		minlev = num
	}
	if scut.Ulev(uni.Dat["_user"]) < minlev {
		return fmt.Errorf("You have no rights to do that.")
	}
	return nil
}

// Wraps SolvePuzzles
// Returns error on go on because one uses this function when wants to explicitly call SolvePuzzles (see comment_insert action of content)
func SolvePuzzlesPath(uni *context.Uni, mod_name, action_name string) error {
	auth_opts, go_on := AuthOpts(uni, mod_name, action_name)
	if go_on {
		return fmt.Errorf("Can't solve puzzles: given action is explicitly ignored.")
	}
	return SolvePuzzles(uni, auth_opts)
}

func puzzleOpt(uni *context.Uni, puzzlename string) (map[string]interface{}, error) {
	puzzle_locate := fmt.Sprintf("user.puzzles.%v", puzzlename)
	puzzle_opt, ok := jsonp.GetM(uni.Opt, puzzle_locate)
	if !ok {
		return nil, fmt.Errorf("Cant find puzzle named %v.", puzzlename)
	}
	return puzzle_opt, nil
}

// Run all the spam protection assigned to the given action - if there is any.
// One can specify a minimum user level for the spam protection task.
// Naturally, if the user is above this level, he must not solve the puzzles.
//
// For further information, see documentation of UserAllowed method.
func SolvePuzzles(uni *context.Uni, auth_options map[string]interface{}) error {
	puzzle_group_i, ok := auth_options["puzzles"]
	if !ok {
		return fmt.Errorf("Can't find puzzle names. Returning, because your system is unsecure.") // We return an error here just to be sure.
	}
	puzzle_group, can_fail := user_model.InterpretPuzzleGroup(puzzle_group_i.([]interface{}))
	failed := 0
	failed_puzzles := []string{}
	fmt.Println(puzzle_group)
	for _, v := range puzzle_group {
		puzzle_opt, err := puzzleOpt(uni, v)
		if err != nil {
			return err
		}
		switch v {
		case "hascash":
			err = solveHashcash(uni, puzzle_opt)
		case "honeypot":
			err = solveHoneypot(uni, puzzle_opt)
		case "timer":
			err = solveTimer(uni, puzzle_opt)
		}
		if err != nil {
			failed_puzzles = append(failed_puzzles, v)
			failed++
		}
		if failed > can_fail {
			return err
			//return fmt.Errorf("Failed more than %v puzzles. Failed: %v.", can_fail, failed_puzzles)
		}
	}
	return nil
}

func solveHoneypot(uni *context.Uni, puzzle_opts map[string]interface{}) error {
	return fmt.Errorf(not_impl)
}

func solveHashcash(uni *context.Uni, puzzle_opts map[string]interface{}) error {
	return fmt.Errorf(not_impl)
}

func solveTimer(uni *context.Uni, puzzle_opts map[string]interface{}) error {
	return user_model.SolveTimer(uni.Secret(), uni.Req.Form, puzzle_opts)
}

// Show puzzles for action. Called as a template function, under the name "show_puzzles".
func ShowPuzzlesPath(uni *context.Uni, mod_name, action_name string) (string, error) {
	auth_opts, go_on := AuthOpts(uni, mod_name, action_name)
	if go_on {
		return "", fmt.Errorf("Can't show puzzles: given action is explicitly ignored.")
	}
	return ShowPuzzles(uni, auth_opts)
}

func ShowPuzzles(uni *context.Uni, auth_options map[string]interface{}) (string, error) {
	puzzle_group_i, ok := auth_options["puzzles"]
	if !ok {
		return "", fmt.Errorf("Can't find puzzle names. Returning because your system is unsecure.") // We return an error here just to be sure.
	}
	puzzle_group, _ := user_model.InterpretPuzzleGroup(puzzle_group_i.([]interface{}))
	var ret string
	for _, v := range puzzle_group {
		puzzle_opt, err := puzzleOpt(uni, v)
		if err != nil {
			return ret, err
		}
		var str string
		switch v {
		case "hascash":
			str, err = showHashcash(uni, puzzle_opt)
		case "timer":
			str, err = showTimer(uni, puzzle_opt)
		case "honeypot":
			str, err = showHoneypot(uni, puzzle_opt)
		case "default":
			str, err = "", fmt.Errorf("Can't find puzzle named %v.", v)
		}
		ret = ret + str
		if err != nil {
			return ret, nil
		}
	}
	return ret, nil
}

func showHashcash(uni *context.Uni, puzzle_opt map[string]interface{}) (string, error) {
	return "", fmt.Errorf(not_impl)
}

func showTimer(uni *context.Uni, puzzle_opt map[string]interface{}) (string, error) {
	return user_model.ShowTimer(uni.Secret(), puzzle_opt)
}

func showHoneypot(uni *context.Uni, puzzle_opt map[string]interface{}) (string, error) {
	return "", fmt.Errorf(not_impl)
}
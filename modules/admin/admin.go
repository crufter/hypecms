// This package implements basic admin functionality.
// - Admin login, or even register if the site has no admin.
// - Installation/uninstallation of modules.
// - Editing of the currently used options document (available under uni.Opts)
// - A view containing links to installed modules.
package admin

import (
	"encoding/json"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/mod"
	"github.com/opesun/hypecms/modules/user"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"io/ioutil"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"path/filepath"
	"strings"
	"time"
)

type m map[string]interface{}

func adErr(uni *context.Uni) {
	if r := recover(); r != nil {
		uni.Put("There was an error running the admin module.\n", r)
	}
}

func SiteHasAdmin(db *mgo.Database) bool {
	var v interface{}
	db.C("users").Find(m{"level": m{"$gt": 299}}).One(&v)
	return v != nil
}

func regUser(db *mgo.Database, post map[string][]string) map[string]interface{} {
	res := map[string]interface{}{}
	pass, pass_ok := post["password"]
	pass_again, pass_again_ok := post["password_again"]
	if !pass_ok || !pass_again_ok || len(pass) < 1 || len(pass_again) < 1 || pass[0] != pass_again[0] {
		res["success"] = false
		res["reason"] = "improper passwords"
	} else {
		a := bson.M{"name": "admin", "level": 300, "password": pass[0]}
		err := db.C("users").Insert(a)
		if err != nil {
			res["success"] = false
			res["reason"] = "name is not unique"
		} else {
			res["success"] = true
		}
	}
	return res
}

// Registering yourself as admin is possible if the site has no admin yet.
func RegAdmin(uni *context.Uni) {
	if SiteHasAdmin(uni.Db) {
		res := map[string]interface{}{}
		res["success"] = false
		res["reason"] = "site already has an admin"
		uni.Dat["_cont"] = res
		return
	}
	uni.Dat["_cont"] = regUser(uni.Db, uni.Req.Form)
}

func RegUser(uni *context.Uni) {
	res := map[string]interface{}{}
	if !requireLev(uni.Dat["_user"], 300) {
		res["success"] = false
		res["reason"] = "no rights"
		uni.Dat["_cont"] = res
		return
	}
	uni.Dat["_cont"] = regUser(uni.Db, uni.Req.Form)
}

func Login(uni *context.Uni) {
	user.Login(uni)
}

func Logout(uni *context.Uni) {
	user.Logout(uni)
}

func Index(uni *context.Uni) {
	uni.Dat["_points"] = []string{"admin/index"}
	adm := map[string]interface{}{}
	if v, ok := uni.Opt["Modules"]; ok {
		if mapi, k := v.(map[string]interface{}); k {
			adm["menu"] = []string{}
			for ind, _ := range mapi {
				adm["menu"] = append(adm["menu"].([]string), ind)
			}
		} else {
			adm["error"] = "Modules in options is not a map[string]interface{}."
		}
	} else {
		adm["error"] = "No module installed."
	}
	uni.Dat["admin"] = adm
}

func EditConfig(uni *context.Uni) {
	uni.Dat["_points"] = []string{"admin/edit-config"}
	adm := map[string]interface{}{}
	delete(uni.Opt, "created")
	v, err := json.MarshalIndent(uni.Opt, "", "    ")
	if err == nil {
		adm["options_json"] = string(v)
	} else {
		adm["error"] = err.Error()
	}
	uni.Dat["admin"] = adm
}

func requireLev(usr interface{}, lev int) bool {
	if val, ok := jsonp.GetI(usr, "level"); ok {
		if val >= lev {
			return true
		}
		return false
	}
	return false
}

func SaveConfig(uni *context.Uni) {
	res := map[string]interface{}{}
	if !requireLev(uni.Dat["_user"], 300) {
		res["success"] = false
		res["reason"] = "no rights"
		uni.Dat["_cont"] = res
		return
	}
	jsonenc, ok := uni.Req.Form["option"]
	if ok {
		if len(jsonenc) == 1 {
			var v interface{}
			json.Unmarshal([]byte(jsonenc[0]), &v)
			if v != nil {
				m := v.(map[string]interface{})
				// Just in case
				delete(m, "_id")
				m["created"] = time.Now().Unix()
				uni.Db.C("options").Insert(m)
				res["success"] = true
			} else {
				res["success"] = false
				res["reason"] = "invalid json"
			}
		} else {
			res["success"] = false
			res["reason"] = "multiple option strings received"
		}
	} else {
		res["success"] = false
		res["reason"] = "no option string received"
	}
	uni.Dat["_cont"] = res
}

// TODO: Highlight already installed packages.
func Install(uni *context.Uni) {
	uni.Dat["_points"] = []string{"admin/install"}
	adm := map[string]interface{}{}
	dirs, err := ioutil.ReadDir(filepath.Join(uni.Root, "/modules"))
	if err == nil {
		modules := []string{}
		for _, v := range dirs {
			if v.IsDir() {
				modules = append(modules, v.Name())
			}
		}
		adm["modules"] = modules
	} else {
		adm["error"] = err.Error()
	}
	uni.Dat["admin"] = adm
}

func Uninstall(uni *context.Uni) {
	uni.Dat["_points"] = []string{"admin/uninstall"}
	uni.Dat["admin"] = map[string]interface{}{}
}

func createCopy(db *mgo.Database) bson.ObjectId {
	var v interface{}
	db.C("options").Find(nil).Sort(bson.M{"created": -1}).Limit(1).One(&v)
	ma := v.(bson.M)
	ma["_id"] = bson.NewObjectId()
	ma["created"] = time.Now().Unix()
	db.C("options").Insert(ma)
	return ma["_id"].(bson.ObjectId)
}

// InstallB handles both installing and uninstalling.
func InstallB(uni *context.Uni) {
	res := map[string]interface{}{}
	if !requireLev(uni.Dat["_user"], 300) {
		res["success"] = false
		res["reason"] = "no rights"
		uni.Dat["_cont"] = res
		return
	}
	mode := ""
	if _, k := uni.Dat["_install"]; k {
		mode = "install"
	} else {
		mode = "uninstall"
	}
	ma, err := routep.Comp("/admin/b/"+mode+"/{modulename}", uni.P)
	if err != nil {
		res["success"] = false
		res["reason"] = "bad url at " + mode
		return
	}
	modn, has := ma["modulename"]
	if !has {
		res["success"] = false
		res["reason"] = "no modulename at " + mode
		return
	}
	if _, already := jsonp.Get(uni.Opt, "Modules."+strings.Title(modn)); mode == "install" && already {
		res["sucess"] = false
		res["reason"] = "Module " + strings.Title(modn) + " is already installed."
	} else if mode == "uninstall" && !already {
		res["sucess"] = false
		res["reason"] = "Module " + strings.Title(modn) + " is not installed."
	} else {
		h := mod.GetHook(modn, strings.Title(mode))
		uni.Dat["_option_id"] = createCopy(uni.Db)
		if h != nil {
			h(uni)
			if _, ok := uni.Dat["_"+mode+"_error"]; ok {
				res["success"] = false
				res["reason"] = uni.Dat["_"+mode+"_reason"]
			} else {
				res["success"] = true
			}
		} else {
			res["success"] = false
			res["reason"] = "Module " + strings.Title(modn) + " does not export the Hook " + strings.Title(mode) + "."
		}
	}
	uni.Dat["_cont"] = res
}

func AD(uni *context.Uni) {
	defer adErr(uni)
	if lev, k := jsonp.Get(uni.Dat, "_user.level"); k == false || lev.(int) < 300 {
		if SiteHasAdmin(uni.Db) {
			uni.Dat["_points"] = []string{"admin/login"}
		} else {
			uni.Dat["_points"] = []string{"admin/regadmin"}
		}
		return
	}
	front, err := routep.Comp("/admin/{module}", uni.P)
	if err == nil { // It should be always
		module, ok := front["module"]
		if !ok {
			module = ""
		}
		switch module {
		case "":
			Index(uni)
		case "edit-config":
			EditConfig(uni)
		case "install":
			Install(uni)
		case "uninstall":
			Uninstall(uni)
		default:
			_, installed := jsonp.Get(uni.Opt, "Modules."+module)
			if installed {
				f := mod.GetHook(module, "AD")
				if f != nil {
					f(uni)
				} else {
					uni.Put("Module ", module, " does not export Admin hook.")
				}
			} else {
				uni.Put("There is no module named ", module, " installed.")
			}
		}
	} else {
		uni.Put("Control is routed to Admin display, but it does not like the url structure.")
	}
}

func AB(uni *context.Uni) {
	m, k := routep.Comp("/admin/b/{action}", uni.P)
	if k == nil {
		switch m["action"] {
		case "regadmin":
			RegAdmin(uni)
		case "reguser":
			RegUser(uni)
		case "adminlogin":
			Login(uni)
		case "logout":
			Logout(uni)
		case "save-config":
			SaveConfig(uni)
		case "install":
			InstallB(uni)
		case "uninstall":
			uni.Dat["_uninstall"] = true
			InstallB(uni)
		default:
			uni.Put("Unknown admin action.")
		}
	} else {
		uni.Put("Control is routed to Admin back, but it does not like the url structure somehow.")
	}
}

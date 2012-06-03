// Ez a package nagyon alap default admin funkcionalitást implementál.
package admin

import (
	"encoding/json"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/mod"
	"github.com/opesun/hypecms/modules/user"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"strings"
	"time"
)

type m map[string]interface{}

func AdErr(uni *context.Uni) {
	if r := recover(); r != nil {
		uni.Put("hiba történt az admin modul futtatása közben", r)
	}
}

func SiteHasAdmin(db *mgo.Database) bool {
	var v interface{}
	db.C("users").Find(m{"level": m{"$gt": 299}}).One(&v)
	return v != nil
}

// Regging yourself as admin is possible if the site has no admin yet.
func RegAdmin(uni *context.Uni) {
	res := map[string]interface{}{}
	if SiteHasAdmin(uni.Db) {
		res["success"] = false
		res["reason"] = "site already has an admin"
		uni.Dat["_cont"] = res
		return
	}
	post := uni.Req.Form
	pass, pass_ok := post["password"]
	pass_again, pass_again_ok := post["password_again"]
	if !pass_ok || !pass_again_ok || len(pass) < 1 || len(pass_again) < 1 || pass[0] != pass_again[0] {
		res["success"] = false
		res["reason"] = "improper passwords"
	} else {
		a := bson.M{"name": "admin", "level": 300, "password": pass[0]}
		err := uni.Db.C("users").Insert(a)
		if err != nil {
			res["success"] = false
			res["reason"] = "name is not unique"
		} else {
			res["success"] = true
		}
	}
	uni.Dat["_cont"] = res
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
		if slice, k := v.([]interface{}); k {
			adm["menu"] = []string{}
			for _, val := range slice {
				adm["menu"] = append(adm["menu"].([]string), val.(string))
			}
		} else {
			adm["error"] = "Modules in options is not a slice"
		}
	} else {
		adm["error"] = "no_module_installed"
	}
	uni.Dat["admin"] = adm
}

func EditConfig(uni *context.Uni) {
	uni.Dat["_points"] = []string{"admin/edit-config"}
	adm := map[string]interface{}{}
	v, err := json.MarshalIndent(uni.Opt, "", "    ")
	if err == nil {
		adm["options_json"] = string(v)
	} else {
		adm["error"] = err.Error()
	}
	uni.Dat["admin"] = adm
}

func SaveConfig(uni *context.Uni) {
	var v interface{}
	uni.Db.C("options").Find(nil).Sort(bson.M{"date": -1}).Limit(1).One(&v)
	if v == nil {
		uni.Put("something is real fucked up")
		return
	}
	m := map[string]interface{}(v.(bson.M))
	delete(m, "_id")
	m["date"] = time.Now().Unix()
	uni.Db.C("options").Insert(m)
	uni.Dat["_cont"] = map[string]interface{}{"success": true}
}

func AD(uni *context.Uni) {
	defer AdErr(uni)
	if lev, k := jsonp.Get(uni.Dat, "_user.level"); k == false || lev.(int) < 300 {
		if SiteHasAdmin(uni.Db) {
			uni.Dat["_points"] = []string{"admin/login"}
		} else {
			uni.Dat["_points"] = []string{"admin/regadmin"}
		}
		return
	}
	front, k := routep.Comp("/admin/{module}", uni.P)
	if k == "" { // It should be always
		module, ok := front["module"]
		if !ok {
			module = ""
		}
		switch module {
		case "":
			Index(uni)
		case "edit-config":
			EditConfig(uni)
		default:
			_, installed := jsonp.Get(uni.Opt, "Modules."+strings.Title(module))
			if installed {
				f := mod.GetHook(module, "Admin")
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
	if k == "" { // admin saját eseményei
		switch m["action"] {
		case "regadmin":
			RegAdmin(uni)
		case "adminlogin":
			Login(uni)
		case "logout":
			Logout(uni)
		case "save-config":
			SaveConfig(uni)
		default:
			uni.Put("Unknown admin action.")
		}
	} else {
		uni.Put("Control is routed to Admin back, but it does not like the url structure.")
	}
}

// This package implements basic admin functionality.
// - Admin login, or even register if the site has no admin.
// - Installation/uninstallation of modules.
// - Editing of the currently used options document (available under uni.Opts)
// - A view containing links to installed modules.
package admin

import (
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/admin/model"
	"github.com/opesun/hypecms/modules/user"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"labix.org/v2/mgo/bson"
	"runtime/debug"
	"strings"
)

type m map[string]interface{}

func adErr(uni *context.Uni) {
	if r := recover(); r != nil {
		uni.Put("There was an error running the admin module.\n", r)
		debug.PrintStack()
	}
}

// Registering yourself as admin is possible if the site has no admin yet.
func RegFirstAdmin(uni *context.Uni) error {
	if admin_model.SiteHasAdmin(uni.Db) {
		return fmt.Errorf("site already has an admin.")
	}
	return admin_model.RegFirstAdmin(uni.Db, map[string][]string(uni.Req.Form))
}

func RegAdmin(uni *context.Uni) error {
	if !requireLev(uni.Dat["_user"], 300) {
		return fmt.Errorf("No rights")
	}
	return admin_model.RegAdmin(uni.Db, uni.Ev, uni.Req.Form)
}

func RegUser(uni *context.Uni) error {
	if !requireLev(uni.Dat["_user"], 300) {
		return fmt.Errorf("No rights")
	}
	return admin_model.RegUser(uni.Db, uni.Ev, uni.Req.Form)
}

func Login(uni *context.Uni) error {
	return user.Login(uni)
}

func Logout(uni *context.Uni) error {
	return user.Logout(uni)
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

func SaveConfig(uni *context.Uni) error {
	if !requireLev(uni.Dat["_user"], 300) {
		return fmt.Errorf("No rights to save config.")
	}
	jsonenc, ok := uni.Req.Form["option"]
	if ok {
		if len(jsonenc) == 1 {
			return admin_model.SaveConfig(uni.Db, uni.Ev, jsonenc[0])
		} else {
			return fmt.Errorf("Multiple option strings received.")
		}
	} else {
		return fmt.Errorf("No option string received.")
	}
	return nil
}

// Install and Uninstall hooks all have the same signature: func(*context.Uni, bson.ObjectId) error
// InstallB handles both installing and uninstalling.
func InstallB(uni *context.Uni, mode string) error {
	if !requireLev(uni.Dat["_user"], 300) {
		return fmt.Errorf("No rights to install or uninstall a module.")
	}
	ma, err := routep.Comp("/admin/b/"+mode+"/{modulename}", uni.P)
	if err != nil {
		return fmt.Errorf("Bad url at " + mode)
	}
	modn, has := ma["modulename"]
	if !has {
		return fmt.Errorf("No modulename at %v.", mode)
	}
	obj_id, ierr := admin_model.InstallB(uni.Db, uni.Ev, uni.Opt, modn, mode)
	if ierr != nil {
		return ierr
	}
	h := uni.GetHook(modn, strings.Title(mode))
	if h == nil {
		return fmt.Errorf("Module %v does not export the Hook %v.", modn, mode)
	}
	hook, ok := h.(func(*context.Uni, bson.ObjectId) error)
	if !ok {
		return fmt.Errorf("%v hook of module %v has bad signature.", mode, modn)
	}
	return hook(uni, obj_id)
}

func AB(uni *context.Uni, action string) error {
	var r error
	switch action {
	case "regfirstadmin":
		r = RegFirstAdmin(uni)
	case "reguser":
		r = RegUser(uni)
	case "adminlogin":
		r = Login(uni)
	case "logout":
		r = Logout(uni)
	case "save-config":
		r = SaveConfig(uni)
	case "install":
		r = InstallB(uni, "install")
	case "uninstall":
		r = InstallB(uni, "uninstall")
	default:
		return fmt.Errorf("Unknown admin action.")
	}
	return r
}

// This package implements basic admin functionality.
// - Admin login, or even register if the site has no admin.
// - Installation/uninstallation of modules.
// - Editing of the currently used options document (available under uni.Opts)
// - A view containing links to installed modules.
package admin

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/mod"
	"github.com/opesun/hypecms/modules/user"
	"github.com/opesun/hypecms/modules/admin/model"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"strings"
	"runtime/debug"
	"fmt"
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
		return fmt.Errorf("site already has an admin")
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
		return fmt.Errorf("No rights to update options collection.")
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

// InstallB handles both installing and uninstalling.
func InstallB(uni *context.Uni, mode string) error {
	if !requireLev(uni.Dat["_user"], 300) {
		return fmt.Errorf("No rights")
	}
	ma, err := routep.Comp("/admin/b/" + mode + "/{modulename}", uni.P)
	if err != nil {
		return fmt.Errorf("Bad url at " + mode)
	}
	modn, has := ma["modulename"]
	if !has {
		return fmt.Errorf("No modulename at " + mode)
	}
	obj_id, ierr := admin_model.InstallB(uni.Db, uni.Ev, uni.Opt, modn, mode)
	if ierr != nil {
		return ierr
	}	
	h := mod.GetHook(modn, strings.Title(mode))
	uni.Dat["_option_id"] = obj_id
	if h != nil {
		inst_err := h(uni)
		if inst_err != nil {
			return inst_err
		}
	} else {
		return fmt.Errorf("Module " + modn + " does not export the Hook " + mode + ".")
	}
	return nil
}

func AB(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
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

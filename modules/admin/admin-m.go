// This package implements basic admin functionality.
// - Admin login, or even register if the site has no admin.
// - Installation/uninstallation of modules.
// - Editing of the currently used options document.
// - A view containing links to installed modules.
package admin

import (
	"fmt"
	"github.com/opesun/hypecms/modules/admin/model"
	"github.com/opesun/hypecms/modules/user"
	"github.com/opesun/extract"
	"strings"
)

// Registering yourself as admin is possible if the site has no admin yet.
func (a *C) RegFirstAdmin() error {
	uni := a.uni
	if admin_model.SiteHasAdmin(uni.Db) {
		return fmt.Errorf("Site already has an admin.")
	}
	return admin_model.RegFirstAdmin(uni.Db, uni.Req.Form)
}

func (a *C) RegAdmin() error {
	return admin_model.RegAdmin(a.uni.Db, a.uni.Req.Form)
}

func (a *C) RegUser() error {
	return admin_model.RegUser(a.uni.Db, a.uni.Req.Form)
}

func (a *C) Login() error {
	u := &user.C{}
	u.Init(a.uni)
	return u.Login()
}

func (a *C) Logout() error {
	u := &user.C{}
	u.Init(a.uni)
	return u.Logout()
}

func (a *C) SaveConfig() error {
	uni := a.uni
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

func (a *C) install(mode string) error {
	uni := a.uni
	dat, err := extract.New(map[string]interface{}{"module":"must"}).Extract(uni.Req.Form)
	if err != nil {
		return err
	}
	modn := dat["module"].(string)
	uni.Dat["_cont"] = map[string]interface{}{"module":modn}
	obj_id, ierr := admin_model.InstallB(uni.Db, uni.Ev, uni.Opt, modn, mode)
	if ierr != nil {
		return ierr
	}
	if !uni.Caller.Has(modn, strings.Title(mode)) {
		return fmt.Errorf("Module %v does not export the Hook %v.", modn, mode) 
	}
	ret_rec := func(e error){
		err = e
	}
	// Install and Uninstall hooks all have the same signature: func (a *A)(bson.ObjectId) error
	uni.Caller.Call(modn, strings.Title(mode), ret_rec, obj_id)
	return err
}

func (a *C) Install() error {
	return a.install("install")
}

func (a *C) Uninstall() error {
	return a.install("uninstall")
}

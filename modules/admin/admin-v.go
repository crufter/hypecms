package admin

import (
	"encoding/json"
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/admin/model"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

func Index(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"admin/index"}
	adm := map[string]interface{}{}
	if v, ok := uni.Opt["Modules"]; ok {
		if mapi, k := v.(map[string]interface{}); k {
			items := []string{}
			for ind, _ := range mapi {
				items = append(items, nameize(ind))
			}
			sort.Strings(items)
			adm["menu"] = items
		} else {
			adm["error"] = "Modules in options is not a map[string]interface{}."
		}
	} else {
		adm["error"] = "No module installed."
	}
	uni.Dat["admin"] = adm
	return nil
}

func EditConfig(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"admin/edit-config"}
	adm := map[string]interface{}{}
	delete(uni.Opt, "created")
	v, err := json.MarshalIndent(uni.Opt, "", "\t")
	if err == nil {
		adm["options_json"] = string(v)
	} else {
		adm["error"] = err.Error()
	}
	uni.Dat["admin"] = adm
	return nil
}

func alreadyInstalled(opt map[string]interface{}, modname string) bool {
	_, ok := jsonp.Get(opt, "Modules." + modname)
	return ok
}

// TODO: Highlight already installed packages.
func Install(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"admin/install"}
	adm := map[string]interface{}{}
	dirs, err := ioutil.ReadDir(filepath.Join(uni.Root, "/modules"))
	if err == nil {
		modules := []string{}
		for _, v := range dirs {
			if v.IsDir() && !alreadyInstalled(uni.Opt, v.Name()) && uni.Caller.Has("hooks", v.Name(), "Install") {	// TODO: this is slow.
				modules = append(modules, nameize(v.Name()))
			}
		}
		sort.Strings(modules)
		adm["modules"] = modules
	} else {
		adm["error"] = err.Error()
	}
	uni.Dat["admin"] = adm
	return nil
}

// Do not turn this on ATM!
// Links in admin index, install and uninstall views will go nuts.
func nameize(s string) string {
	// s = strings.Replace(s, "_", " ", -1)
	// return strings.Title(s)
	return s
}

func Uninstall(uni *context.Uni) error {
	installed_mods := []string{}
	modules, has := uni.Opt["Modules"]
	if has {
		for i, _ := range modules.(map[string]interface{}) {
			installed_mods = append(installed_mods, nameize(i))			// TODO: what to do with modules not having a proper Uninstall hook?
		}
	}
	sort.Strings(installed_mods)
	uni.Dat["installed_modules"] = installed_mods
	uni.Dat["_points"] = []string{"admin/uninstall"}
	return nil
}

func Viewnameize(viewname string) string {
	viewname = strings.Replace(viewname, "-", " ", -1)
	viewname = strings.Title(viewname)
	return strings.Replace(viewname, " ", "", -1)
}

func AD(uni *context.Uni) error {
	defer adErr(uni)
	var err error
	if lev, k := jsonp.Get(uni.Dat, "_user.level"); k == false || lev.(int) < 300 {
		if admin_model.SiteHasAdmin(uni.Db) {
			uni.Dat["_points"] = []string{"admin/login"}
		} else {
			uni.Dat["_points"] = []string{"admin/regfirstadmin"}
		}
		return nil
	}
	m, cerr := routep.Comp("/admin/{modname}", uni.P)
	if cerr != nil { // It should be always nil anyway.
		return fmt.Errorf("Control is routed to Admin display, but it does not like the url structure.")
	}
	modname, _ := m["modname"]
	switch modname {
	case "":
		err = Index(uni)
	case "edit-config":
		err = EditConfig(uni)
	case "install":
		err = Install(uni)
	case "uninstall":
		err = Uninstall(uni)
	default:
		_, installed := jsonp.Get(uni.Opt, "Modules."+modname)
		if !installed {
			err = fmt.Errorf("There is no module named ", modname, " installed.")
		}
		var viewname string
		if len(uni.Paths) < 4 {
			viewname = "index"
		} else {
			viewname = uni.Paths[3]
		}
		uni.Caller.Call("views", modname, "AdminInit", nil)
		sanitized_viewname := Viewnameize(viewname)
		if !uni.Caller.Has("views", modname, sanitized_viewname) {
			err = fmt.Errorf("Module %v has no view named %v.", modname, sanitized_viewname)
		}
		ret_rec := func(e error) {
			err = e
		}
		uni.Dat["_points"] = []string{modname+"/"+viewname}
		uni.Caller.Call("views", modname, sanitized_viewname, ret_rec)
	}
	return err
}

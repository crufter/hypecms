package admin

import(
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
	"github.com/opesun/hypecms/api/mod"
	"sort"
	"fmt"
)

func Index(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"admin/index"}
	adm := map[string]interface{}{}
	if v, ok := uni.Opt["Modules"]; ok {
		if mapi, k := v.(map[string]interface{}); k {
			items := []string{}
			for ind, _ := range mapi {
				items = append(items, ind)
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
	v, err := json.MarshalIndent(uni.Opt, "", "    ")
	if err == nil {
		adm["options_json"] = string(v)
	} else {
		adm["error"] = err.Error()
	}
	uni.Dat["admin"] = adm
	return nil
}

// TODO: Highlight already installed packages.
func Install(uni *context.Uni) error {
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
	return nil
}

func Uninstall(uni *context.Uni) error {
	installed_mods := []string{}
	modules, has := uni.Opt["Modules"]
	if has {
		for i, _ := range modules.(map[string]interface{}) {
			installed_mods = append(installed_mods, i)
		}
	}
	uni.Dat["installed_modules"] = installed_mods
	uni.Dat["_points"] = []string{"admin/uninstall"}
	return nil
}

func AD(uni *context.Uni) error {
	defer adErr(uni)
	var err error
	if lev, k := jsonp.Get(uni.Dat, "_user.level"); k == false || lev.(int) < 300 {
		if SiteHasAdmin(uni.Db) {
			uni.Dat["_points"] = []string{"admin/login"}
		} else {
			uni.Dat["_points"] = []string{"admin/regadmin"}
		}
		return nil
	}
	m, cerr := routep.Comp("/admin/{modname}", uni.P)
	if cerr == nil { // It should be always nil anyway.
		modname, ok := m["modname"]
		if !ok {
			modname = ""
		}
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
			_, installed := jsonp.Get(uni.Opt, "Modules." + modname)
			if installed {
				f := mod.GetHook(modname, "AD")
				if f != nil {
					err = f(uni)
				} else {
					err = fmt.Errorf("Module ", modname, " does not export hook AD.")
				}
			} else {
				err = fmt.Errorf("There is no module named ", modname, " installed.")
			}
		}
	} else {
		err = fmt.Errorf("Control is routed to Admin display, but it does not like the url structure.")
	}
	return err
}
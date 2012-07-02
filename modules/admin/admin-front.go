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
)

func Index(uni *context.Uni) {
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
	installed_mods := []string{}
	modules, has := uni.Opt["Modules"]
	if has {
		for i, _ := range modules.(map[string]interface{}) {
			installed_mods = append(installed_mods, i)
		}
	}
	uni.Dat["installed_modules"] = installed_mods
	uni.Dat["_points"] = []string{"admin/uninstall"}
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
	m, err := routep.Comp("/admin/{modname}", uni.P)
	if err == nil { // It should be always nil anyway.
		modname, ok := m["modname"]
		if !ok {
			modname = ""
		}
		switch modname {
		case "":
			Index(uni)
		case "edit-config":
			EditConfig(uni)
		case "install":
			Install(uni)
		case "uninstall":
			Uninstall(uni)
		default:
			_, installed := jsonp.Get(uni.Opt, "Modules." + modname)
			if installed {
				f := mod.GetHook(modname, "AD")
				if f != nil {
					f(uni)
				} else {
					uni.Put("Module ", modname, " does not export hook AD.")
				}
			} else {
				uni.Put("There is no module named ", modname, " installed.")
			}
		}
	} else {
		uni.Put("Control is routed to Admin display, but it does not like the url structure.")
	}
}
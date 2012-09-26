package admin

import (
	"encoding/json"
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/jsonp"
	"io/ioutil"
	"path/filepath"
	"sort"
)

func (v *V) Index() error {
	uni := v.uni
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

func (v *V) EditConfig() error {
	uni := v.uni
	uni.Dat["_points"] = []string{"admin/edit-config"}
	adm := map[string]interface{}{}
	delete(uni.Opt, "created")
	marsh, err := json.MarshalIndent(uni.Opt, "", "\t")
	if err == nil {
		adm["options_json"] = string(marsh)
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
func (v *V) Install() error {
	uni := v.uni
	uni.Dat["_points"] = []string{"admin/install"}
	adm := map[string]interface{}{}
	dirs, err := ioutil.ReadDir(filepath.Join(uni.Root, "/modules"))
	if err == nil {
		modules := []string{}
		for _, val := range dirs {
			if val.IsDir() && !alreadyInstalled(uni.Opt, val.Name()) && uni.Caller.Has("hooks", val.Name(), "Install") {	// TODO: this is slow.
				modules = append(modules, nameize(val.Name()))
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

func (v *V) Uninstall() error {
	uni := v.uni
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

type V struct{
	uni *context.Uni
}

func Views(uni *context.Uni) *V {
	return &V{uni}
}

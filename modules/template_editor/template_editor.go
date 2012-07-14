// Package template_editor implements a minimalistic but idiomatic plugin for HypeCMS.
package template_editor

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/scut"
	//"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"labix.org/v2/mgo/bson"
	"io/ioutil"
	"fmt"
	"path/filepath"
	"strings"
	"github.com/opesun/hypecms/modules/template_editor/model"
)

// mod.GetHook accesses certain functions dynamically trough this.
var Hooks = map[string]func(*context.Uni) error {
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}
// main.runBackHooks invokes this trough mod.GetHook.
func Back(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
	switch action {
	// You can dispatch your background operations here.
	}
	return nil
}

func isDir(filep string) bool {
	filep_s := strings.Split(filep, "/")
	if strings.Index(filep_s[len(filep_s)-1], ".") == -1 {
		return true
	}
	return false
}

type Breadc struct {
	Name string
	Path string
}

func createBreadCrumb(fs []string) []Breadc {
	ret := []Breadc{}
	for i:=1; i<len(fs); i++ {
		ret = append(ret, Breadc{fs[i], "/" + filepath.Join(fs[:i+1]...)})
	}
	return ret
}

func View(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"template_editor/view"}
	filepath_s, has := uni.Req.Form["file"]
	if !has {
		uni.Dat["error"] = "Can't find file parameter."
		return nil
	}
	filepath_str := filepath_s[0]
	tpath := scut.GetTPath(uni.Opt)
	uni.Dat["breadcrumb"] = createBreadCrumb(strings.Split(filepath_str, "/"))
	uni.Dat["filepath"] = filepath.Join(tpath, filepath_str)
	uni.Dat["raw_path"] = filepath_str
	if isDir(filepath_str) {
		fileinfos, read_err := ioutil.ReadDir(filepath.Join(uni.Root, tpath, filepath_str))
		if read_err != nil {
			uni.Dat["error"] = read_err.Error()
		}
		uni.Dat["dir"] = fileinfos
	} else {
		file_b, read_err := ioutil.ReadFile(filepath.Join(uni.Root, tpath, filepath_str))
		if read_err != nil {
			uni.Dat["error"] = "Can't find specified file."
			return nil
		}
		uni.Dat["file"] = string(file_b)
	}
	return nil
}

func Index(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"template_editor/index"}
	return nil
}

// admin.AD invokes this trough mod.GetHook.
func AD(uni *context.Uni) error {
	ma, err := routep.Comp("/admin/template_editor/{view}", uni.P)
	if err != nil { return err }
	var r error
	switch ma["view"] {
		case "":
			r = Index(uni)
		case "view":
			r = View(uni)
		default:
			return fmt.Errorf("Unkown view at template_editor admin.")
	}
	return r
}

// admin.Install invokes this trough mod.GetHook.
func Install(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	return template_editor_model.Install(uni.Db, id)
}

// Admin Install invokes this trough mod.GetHook.
func Uninstall(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	return template_editor_model.Uninstall(uni.Db, id)
}

// main.runDebug invokes this trough mod.GetHook.
func Test(uni *context.Uni) error {
	return nil
}

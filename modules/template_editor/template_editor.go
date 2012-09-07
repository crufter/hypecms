// TODO: make the difference between current and noncurrent template browsing/editing disappear, so the code can get simpler and easier to read/develope.
// (Background operations will need some unnecessary parameters then? Rethink.)
package template_editor

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/scut"
	//"github.com/opesun/jsonp"
	"fmt"
	te_model "github.com/opesun/hypecms/modules/template_editor/model"
	"github.com/opesun/routep"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"strings"
)

// mod.GetHook accesses certain functions dynamically trough this.
var Hooks = map[string]interface{}{
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}

func NewFile(uni *context.Uni) error {
	return te_model.NewFile(uni.Opt, map[string][]string(uni.Req.Form), uni.Root, uni.Req.Host)
}

func SaveFile(uni *context.Uni) error {
	return te_model.SaveFile(uni.Opt, map[string][]string(uni.Req.Form), uni.Root, uni.Req.Host)
}

func DeleteFile(uni *context.Uni) error {
	return te_model.DeleteFile(uni.Opt, map[string][]string(uni.Req.Form), uni.Root, uni.Req.Host)
}

func ForkPublic(uni *context.Uni) error {
	return te_model.ForkPublic(uni.Db, uni.Opt, uni.Root, uni.Req.Host)
}

func PublishPrivate(uni *context.Uni) error {
	return te_model.PublishPrivate(uni.Db, uni.Opt, map[string][]string(uni.Req.Form), uni.Root, uni.Req.Host)
}

func DeletePrivate(uni *context.Uni) error {
	return te_model.DeletePrivate(uni.Opt, map[string][]string(uni.Req.Form), uni.Root, uni.Req.Host)
}

func ForkPrivate(uni *context.Uni) error {
	return te_model.ForkPrivate(uni.Db, uni.Opt, map[string][]string(uni.Req.Form), uni.Root, uni.Req.Host)
}

func SwitchToTemplate(uni *context.Uni) error {
	return te_model.SwitchToTemplate(uni.Db, map[string][]string(uni.Req.Form))
}

// main.runBackHooks invokes this trough mod.GetHook.
func Back(uni *context.Uni, action string) error {
	if scut.NotAdmin(uni.Dat["_user"]) {
		return fmt.Errorf("You have no rights to do that.")
	}
	var r error
	switch action {
	case "new_file":
		r = NewFile(uni)
	case "save_file":
		r = SaveFile(uni)
	case "delete_file":
		r = DeleteFile(uni)
	case "fork_public":
		r = ForkPublic(uni)
	case "publish_private":
		r = PublishPrivate(uni)
	case "delete_private":
		r = DeletePrivate(uni)
	case "fork_private":
		r = ForkPrivate(uni)
	case "switch_to_template":
		r = SwitchToTemplate(uni)
	default:
		return fmt.Errorf("Unkown action at template_editor.")
	}
	return r
}

func threePath(host, typ, name string) (string, error) {
	var ret string
	switch typ {
	case "mod":
		ret = filepath.Join("modules", name, "tpl")
	case "public":
		ret = filepath.Join("templates", "public", name)
	case "private":
		ret = filepath.Join("templates", "private", host, name)
	default:
		return "", fmt.Errorf("Unkown template type.")
	}
	return ret, nil
}

func canMod(typ string) bool {
	return typ == "private"
}

func view(current bool, opt map[string]interface{}, root, host, typ, name, filepath_str string) map[string]interface{} {
	ret := map[string]interface{}{}
	tpath, path_err := threePath(host, typ, name)
	if path_err != nil {
		ret["error"] = path_err
		return ret
	}
	ret["template_name"] = name
	ret["breadcrumb"] = te_model.CreateBreadCrumb(strings.Split(filepath_str, "/"))
	ret["can_modify"] = canMod(typ)
	ret["current"] = current
	ret["typ"] = typ
	if typ == "mod" {
		ret["is_mod"] = true
	}
	ret["filepath"] = filepath.Join(tpath, filepath_str)
	ret["raw_path"] = filepath_str
	if te_model.IsDir(filepath_str) {
		fileinfos, read_err := ioutil.ReadDir(filepath.Join(root, tpath, filepath_str))
		if read_err != nil {
			ret["error"] = read_err.Error()
			return ret
		}
		ret["dir"] = fileinfos
		ret["is_dir"] = true
	} else {
		file_b, read_err := ioutil.ReadFile(filepath.Join(root, tpath, filepath_str))
		if read_err != nil {
			ret["error"] = "Can't find specified file."
			return ret
		}
		if len(file_b) == 0 {
			ret["file"] = "[Empty file.]" // A temporary hack, because the codemirror editor is not displayed when editing an empty file. It is definitely a client-side javascript problem.
		} else {
			ret["included"] = te_model.ReqLinks(opt, string(file_b), root, host)
			ret["file"] = string(file_b)
		}
	}
	return ret
}

// Get parameters:
// type: tpl, private, public
// file: path of file
// name: name of template or module
func View(uni *context.Uni, typ, name string) error {
	uni.Dat["_points"] = []string{"template_editor/view"}
	filepath_s, has := uni.Req.Form["file"]
	if !has {
		uni.Dat["error"] = "Can't find file parameter."
		return nil
	}
	no_typ := len(typ) == 0
	no_name := len(name) == 0
	if no_typ && no_name {
		scut.Merge(uni.Dat, view(true, uni.Opt, uni.Root, uni.Req.Host, scut.TemplateType(uni.Opt), scut.TemplateName(uni.Opt), filepath_s[0]))
		return nil
	}
	if no_typ {
		uni.Dat["error"] = "Got no template name."
		return nil
	}
	if no_name {
		uni.Dat["error"] = "Got no template type."
		return nil
	}
	scut.Merge(uni.Dat, view(false, uni.Opt, uni.Root, uni.Req.Host, typ, name, filepath_s[0]))
	return nil
}

func Index(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"template_editor/index"}
	uni.Dat["template_name"] = scut.TemplateName(uni.Opt)
	uni.Dat["can_modify"] = te_model.CanModifyTemplate(uni.Opt)
	return nil
}

func search(uni *context.Uni, path string) error {
	fileinfos, read_err := ioutil.ReadDir(filepath.Join(uni.Root, path))
	if read_err != nil {
		uni.Dat["error"] = "Cant read path " + path
		return nil
	}
	term_s, has := uni.Req.Form["search"]
	term := ""
	if has {
		term = term_s[0]
	}
	uni.Dat["dir"] = te_model.Contains(fileinfos, term)
	return nil
}

func SearchPublic(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"template_editor/search"}
	uni.Dat["is_public"] = true
	path := filepath.Join("templates", "public")
	search(uni, path)
	return nil
}

func SearchPrivate(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"template_editor/search"}
	uni.Dat["is_private"] = true
	path := filepath.Join("templates", "private", uni.Req.Host)
	search(uni, path)
	return nil
}

func SearchMod(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"template_editor/search"}
	uni.Dat["is_mod"] = true
	path := filepath.Join("modules")
	search(uni, path)
	return nil
}

// admin.AD invokes this trough mod.GetHook.
func AD(uni *context.Uni) error {
	ma, err := routep.Comp("/admin/template_editor/{view}/{typ}/{name}", uni.P)
	if err != nil {
		return err
	}
	var r error
	switch ma["view"] {
	case "":
		r = Index(uni)
	case "view":
		r = View(uni, ma["typ"], ma["name"])
	case "search-public":
		r = SearchPublic(uni)
	case "search-private":
		r = SearchPrivate(uni)
	case "search-mod":
		r = SearchMod(uni)
	default:
		return fmt.Errorf("Unkown view at template_editor admin.")
	}
	return r
}

// admin.Install invokes this trough mod.GetHook.
func Install(uni *context.Uni, id bson.ObjectId) error {
	return te_model.Install(uni.Db, id)
}

// Admin Install invokes this trough mod.GetHook.
func Uninstall(uni *context.Uni, id bson.ObjectId) error {
	return te_model.Uninstall(uni.Db, id)
}

// main.runDebug invokes this trough mod.GetHook.
func Test(uni *context.Uni) error {
	return nil
}

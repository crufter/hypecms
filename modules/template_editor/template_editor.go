// TODO: make the difference between current and noncurrent template browsing/editing disappear, so the code can get simpler and easier to read/develope.
// (Background operations will need some unnecessary parameters then? Rethink.)
package template_editor

import (
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/scut"
	te_model "github.com/opesun/hypecms/modules/template_editor/model"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"strings"
)

func (a *A) NewFile() error {
	uni := a.uni
	return te_model.NewFile(uni.Opt, uni.Req.Form, uni.Root, uni.Req.Host)
}

func (a *A) SaveFile() error {
	uni := a.uni
	return te_model.SaveFile(uni.Opt, uni.Req.Form, uni.Root, uni.Req.Host)
}

func (a *A) DeleteFile() error {
	uni := a.uni
	return te_model.DeleteFile(uni.Opt, uni.Req.Form, uni.Root, uni.Req.Host)
}

func (a *A) ForkPublic() error {
	uni := a.uni
	return te_model.ForkPublic(uni.Db, uni.Opt, uni.Root, uni.Req.Host)
}

func (a *A) PublishPrivate() error {
	uni := a.uni
	return te_model.PublishPrivate(uni.Db, uni.Opt, uni.Req.Form, uni.Root, uni.Req.Host)
}

func (a *A) DeletePrivate() error {
	uni := a.uni
	return te_model.DeletePrivate(uni.Opt, uni.Req.Form, uni.Root, uni.Req.Host)
}

func (a *A) ForkPrivate() error {
	uni := a.uni
	return te_model.ForkPrivate(uni.Db, uni.Opt, uni.Req.Form, uni.Root, uni.Req.Host)
}

func (a *A) SwitchToTemplate() error {
	uni := a.uni
	return te_model.SwitchToTemplate(uni.Db, uni.Req.Form, uni.Root, uni.Req.Host)
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

func (v *V) view(current bool, filepath_str string) map[string]interface{} {
	uni := v.uni
	opt := uni.Opt
	root :=uni.Root
	host := uni.Req.Host
	typ := scut.TemplateType(uni.Opt)
	name := scut.TemplateName(uni.Opt)
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
func (v *V) View() error {
	uni := v.uni
	var typ, name string
	if val, has := uni.Req.Form["type"]; has {
		typ = val[0]
	}
	if val, has := uni.Req.Form["name"]; has {
		name = val[0]
	}
	filepath_s, has := uni.Req.Form["file"]
	if !has {
		uni.Dat["error"] = "Can't find file parameter."
		return nil
	}
	no_typ := len(typ) == 0
	no_name := len(name) == 0
	if no_typ && no_name {
		scut.Merge(uni.Dat, v.view(true, filepath_s[0]))
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
	scut.Merge(uni.Dat, v.view(false, filepath_s[0]))
	return nil
}

func (v *V) Index() error {
	uni := v.uni
	uni.Dat["template_name"] = scut.TemplateName(uni.Opt)
	uni.Dat["can_modify"] = te_model.CanModifyTemplate(uni.Opt)
	return nil
}

func (v *V) search(path string) error {
	uni := v.uni
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

// Search amongst public templates.
func (v *V) SearchPublic() error {
	uni := v.uni
	uni.Dat["is_public"] = true
	path := filepath.Join("templates", "public")
	v.search(path)
	uni.Dat["_points"] = []string{"template_editor/search"}
	return nil
}

// Search amongst private templates.
func (v *V) SearchPrivate() error {
	uni := v.uni
	uni.Dat["is_private"] = true
	path := filepath.Join("templates", "private", uni.Req.Host)
	v.search(path)
	uni.Dat["_points"] = []string{"template_editor/search"}
	return nil
}

// Search amongst modules.
func (v *V) SearchMod() error {
	uni := v.uni
	uni.Dat["is_mod"] = true
	path := filepath.Join("modules")
	v.search(path)
	uni.Dat["_points"] = []string{"template_editor/search"}
	return nil
}

// admin.Install invokes this trough mod.GetHook.
func (h *H) Install(id bson.ObjectId) error {
	return te_model.Install(h.uni.Db, id)
}

// Admin Install invokes this trough mod.GetHook.
func (h *H) Uninstall(id bson.ObjectId) error {
	return te_model.Uninstall(h.uni.Db, id)
}

type A struct {
	uni *context.Uni
}

func Actions(uni *context.Uni) *A {
	return &A{uni}
}

type H struct {
	uni *context.Uni
}

func Hooks(uni *context.Uni) *H {
	return &H{uni}
}

type V struct {
	uni *context.Uni
}

func Views(uni *context.Uni) *V {
	return &V{uni}
}
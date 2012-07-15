package template_editor_model

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/api/scut"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/opesun/copyrecur"
	"github.com/opesun/hypecms/model/basic"
)

type m map[string]interface{}

const(
	cant_mod_public = "Can't modify public template."
)

func CanModifyTemplate(opt map[string]interface{}) bool {
	if scut.TemplateType(opt) == "public" {
		return false
	}
	return true
}

func NewFile(opt map[string]interface{}, inp map[string][]string, root, host string) error {
	if !CanModifyTemplate(opt) {
		return fmt.Errorf(cant_mod_public)
	}
	rule := map[string]interface{}{
		"filepath": "must",
	}
	dat, e_err := extract.New(rule).Extract(inp)
	if e_err != nil {
		return e_err
	}
	fp := dat["filepath"].(string)
	return ioutil.WriteFile(filepath.Join(root, scut.GetTPath(opt, host), fp), []byte(""), os.ModePerm)
}

func SaveFile(opt map[string]interface{}, inp map[string][]string, root, host string) error {
	if !CanModifyTemplate(opt) {
		return fmt.Errorf(cant_mod_public)
	}
	rule := map[string]interface{}{
		"filepath": "must",
		"content":	"must",
	}
	dat, e_err := extract.New(rule).Extract(inp)
	if e_err != nil {
		return e_err
	}
	fp := dat["filepath"].(string)
	content := dat["content"].(string)
	return ioutil.WriteFile(filepath.Join(root, scut.GetTPath(opt, host), fp), []byte(content), os.ModePerm)
}

func DeleteFile(opt map[string]interface{}, inp map[string][]string, root, host string) error {
	if !CanModifyTemplate(opt) {
		return fmt.Errorf(cant_mod_public)
	}
	rule := map[string]interface{}{
		"filepath": "must",
	}
	dat, e_err := extract.New(rule).Extract(inp)
	if e_err != nil {
		return e_err
	}
	fp := dat["filepath"].(string)
	return os.Remove(filepath.Join(root, scut.GetTPath(opt, host), fp))
}

func ForkPublic(db *mgo.Database, opt map[string]interface{}, host, root string) error {
	if CanModifyTemplate(opt) {
		return fmt.Errorf("Template is already private.")
	}
	from := filepath.Join(root, scut.GetTPath(opt, host))
	to := filepath.Join(root, "templates", "private", host, scut.TemplateName(opt))
	copy_err := copyrecur.CopyDir(from, to)
	if copy_err != nil && copy_err.Error() != "Destination already exists" {
		return copy_err
	}
	id := basic.CreateOptCopy(db)
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"TplIsPrivate": true,
		},
	}
	return db.C("options").Update(q, upd)
}

func Install(db *mgo.Database, id bson.ObjectId) error {
	template_editor_options := m{
		// "example": "any value",
	}
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"Modules.template_editor": template_editor_options,
		},
	}
	return db.C("options").Update(q, upd)
}

func Uninstall(db *mgo.Database, id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$unset": m{
			"Modules.template_editor": 1,
		},
	}
	return db.C("options").Update(q, upd)
}
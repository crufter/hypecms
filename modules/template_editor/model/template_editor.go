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
	"strings"
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

func IsDir(filep string) bool {
	filep_s := strings.Split(filep, "/")
	if strings.Index(filep_s[len(filep_s)-1], ".") == -1 {
		return true
	}
	return false
}

// New file OR dir. Filenames without extensions became dirs. RETHINK: This way we lose the ability to create files without extensions.
// Only accessed member of opt will be "TplIsPrivate" in scut.GetTPath. TODO: this is ugly.
func NewFile(opt map[string]interface{}, inp map[string][]string, root, host string) error {
	if !CanModifyTemplate(opt) {
		return fmt.Errorf(cant_mod_public)
	}
	rule := map[string]interface{}{
		"filepath": "must",
		"where":	"must",
	}
	dat, e_err := extract.New(rule).Extract(inp)
	if e_err != nil {
		return e_err
	}
	fp := dat["filepath"].(string)
	where := dat["where"].(string)
	if IsDir(fp) {
		return os.Mkdir(filepath.Join(root, scut.GetTPath(opt, host), where, fp), os.ModePerm)
	}
	return ioutil.WriteFile(filepath.Join(root, scut.GetTPath(opt, host), where, fp), []byte(""), os.ModePerm)
}

// Save an existing file.
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

// Delete a file OR dir. See NewFile for controversy about filenames and extensions.
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
	full_p := filepath.Join(root, scut.GetTPath(opt, host), fp)
	var err error
	if IsDir(fp) {
		err = os.RemoveAll(full_p)
	} else {
		err = os.Remove(full_p)
	}
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("Can't delete file or directory. Probably it does not exist.")
	}
	return nil
}

// Forks a public template into a private one: creates a deep recursive copy of the whole directory tree, so the user
// can edit his own template files as he wishes.
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
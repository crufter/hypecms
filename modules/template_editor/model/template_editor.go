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
	return scut.TemplateType(opt) == "private"
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

// Taken from http://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

// Publish a private template, so others can use it too.
func PublishPrivate(db *mgo.Database, opt map[string]interface{}, inp map[string][]string, host, root string) error {
	rule := map[string]interface{}{
		"public_name": 	"must",
	}
	dat, ex_err := extract.New(rule).Extract(inp)
	if ex_err != nil { return ex_err }
	public_name := dat["public_name"].(string)
	from := filepath.Join(root, "templates", "private", host, scut.TemplateName(opt))
	to := filepath.Join(root, "templates", "public", public_name)
	// copyrecur.CopyDir checks for existence too, but for safety reasons we check here in case copyrecur semantics change.
	exis, exis_err := exists(to)
	if exis {
		return fmt.Errorf("Public template with name " + public_name + " already exists.")
	}
	if exis_err != nil { return exis_err }
	copy_err := copyrecur.CopyDir(from, to)
	if copy_err != nil {
		return fmt.Errorf("There was an error while copying.")
	}
	return nil
}

func Contains(fi []os.FileInfo, term string) []os.FileInfo {
	ret_fis := []os.FileInfo{}
	for _, v := range fi {
		if len(term) == 0 || strings.Index(v.Name(), term) != -1 {
			ret_fis = append(ret_fis, v)
		}
	}
	return ret_fis
}

func Search(root, host, typ, search_str string) ([]os.FileInfo, error) {
	var path string
	if typ == "public" {
		path =  filepath.Join(root, "templates", "public")
	} else {
		path = filepath.Join(root, "templates", "private", host)
	}
	fileinfos, read_err := ioutil.ReadDir(filepath.Join(root, path))
	if read_err != nil { return nil, read_err }
	return Contains(fileinfos, search_str), nil
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
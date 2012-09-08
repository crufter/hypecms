package template_editor_model

import (
	"fmt"
	"github.com/opesun/copyrecur"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/require"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"strings"
)

type m map[string]interface{}

const (
	cant_mod_public = "Can't modify public template."
)

func CanModifyTemplate(opt map[string]interface{}) bool {
	return scut.TemplateType(opt) == "private"
}

// Returns true if a given filepath (relative or absolute) identifies a directory.
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
		"where":    "must",
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
		"content":  "must",
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
func ForkPublic(db *mgo.Database, opt map[string]interface{}, root, host string) error {
	if CanModifyTemplate(opt) {
		return fmt.Errorf("Template is already private.")
	}
	from := filepath.Join(root, scut.GetTPath(opt, host))
	to := filepath.Join(root, "templates", "private", host, scut.TemplateName(opt))
	copy_err := copyrecur.CopyDir(from, to)
	if copy_err != nil {	// && copy_err.Error() != "Destination already exists"
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
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Publish a private template, so others can use it too.
// Copies the whole directory of /templates/private/{host}/{current_template} to /templates/public/{input:public_name}
// Fails if a public template with the chosen name already exists.
func PublishPrivate(db *mgo.Database, opt map[string]interface{}, inp map[string][]string, root, host string) error {
	if scut.TemplateType(opt) == "public" {
		return fmt.Errorf("You can't publish your current template, because it is already public.")
	}
	rule := map[string]interface{}{
		"public_name": map[string]interface{}{
			"must": 1,
			"type": "string",
			"min":	2,
		},
	}
	dat, ex_err := extract.New(rule).Extract(inp)
	if ex_err != nil {
		return ex_err
	}
	public_name := dat["public_name"].(string)
	from := filepath.Join(root, "templates", "private", host, scut.TemplateName(opt))
	to := filepath.Join(root, "templates", "public", public_name)
	// copyrecur.CopyDir checks for existence too, but for safety reasons we check here in case copyrecur semantics change.
	exis, exis_err := Exists(to)
	if exis {
		return fmt.Errorf("Public template with name " + public_name + " already exists.")
	}
	if exis_err != nil {
		return exis_err
	}
	copy_err := copyrecur.CopyDir(from, to)
	if copy_err != nil {
		return fmt.Errorf("There was an error while copying.")
	}
	return nil
}

// Filter
func Contains(fi []os.FileInfo, term string) []os.FileInfo {
	ret_fis := []os.FileInfo{}
	for _, v := range fi {
		if len(term) == 0 || strings.Index(v.Name(), term) != -1 {
			ret_fis = append(ret_fis, v)
		}
	}
	return ret_fis
}

// Delete a whole private template.
func DeletePrivate(opt map[string]interface{}, inp map[string][]string, root, host string) error {
	rule := map[string]interface{}{
		"template_name": "must",
	}
	dat, e_err := extract.New(rule).Extract(inp)
	if e_err != nil {
		return e_err
	}
	template_name := dat["template_name"].(string)
	if template_name == scut.TemplateName(opt) {
		return fmt.Errorf("For safety reasons you can only delete private templates not in use.")
	}
	full_p := filepath.Join(root, "templates", "private", host, template_name)
	err := os.RemoveAll(full_p)
	if err != nil {
		return fmt.Errorf("Can't delete private template named %v. It probably does not exist.", template_name)
	}
	return nil
}

// Fork current private template into an other private one.
// Copies the whole directory from /templates/private/{host}/{current_template} to /templates/private/{host}/{inp:new_private_name}
func ForkPrivate(db *mgo.Database, opt map[string]interface{}, inp map[string][]string, root, host string) error {
	if scut.TemplateType(opt) != "private" {
		return fmt.Errorf("Your current template is not a private one.") // Kinda unsensical error message but ok...
	}
	rule := map[string]interface{}{
		"new_template_name": "must",
	}
	dat, e_err := extract.New(rule).Extract(inp)
	if e_err != nil {
		return e_err
	}
	new_template_name := dat["new_template_name"].(string)
	to := filepath.Join(root, "templates", "private", host, new_template_name)
	e, e_err := Exists(to)
	if e_err != nil {
		return fmt.Errorf("Can't determine if private template exists.")
	}
	if e {
		return fmt.Errorf("Private template named %v already exists.", new_template_name)
	}
	from := filepath.Join(root, "templates", "private", host, scut.TemplateName(opt))
	copy_err := copyrecur.CopyDir(from, to)
	if copy_err != nil {
		return fmt.Errorf("There was an error while copying.")
	}
	id := basic.CreateOptCopy(db)
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"Template": new_template_name,
		},
	}
	return db.C("options").Update(q, upd)
}

// Switches from one template to another.
// Fails if the template we want to switch does not exist.
func SwitchToTemplate(db *mgo.Database, inp map[string][]string, root, host string) error {
	rule := map[string]interface{}{
		"template_name": "must",
		"template_type": "must",
	}
	dat, e_err := extract.New(rule).Extract(inp)
	if e_err != nil {
		return e_err
	}
	template_type := dat["template_type"].(string)
	template_name := dat["template_name"].(string)
	var e bool
	var err error
	if template_type == "public" {
		e, err = Exists(filepath.Join(root, "templates/public", template_name))
	} else {
		e, err = Exists(filepath.Join(root, "templates/private", host, template_name))
	}
	if err != nil {
		return fmt.Errorf("Can't determine if template exists.")
	}
	if !e {
		return fmt.Errorf("%v template named %v does not exist.", template_type, template_name)
	}
	return switchToTemplateDb(db, template_type, template_name)
}

// Does the database operation involved in template switch.
func switchToTemplateDb(db *mgo.Database, template_type, template_name string) error {
	id := basic.CreateOptCopy(db)
	q := m{"_id": id}
	var upd m
	if template_type == "private" {
		upd = m{
			"$set": m{
				"Template":     template_name,
				"TplIsPrivate": true,
			},
		}
	} else {
		upd = m{
			"$set": m{
				"Template": template_name,
			},
			"$unset": m{
				"TplIsPrivate": 1,
			},
		}
	}
	return db.C("options").Update(q, upd)
}

type ReqLink struct {
	Typ      string
	Tempname string
	Filepath string
}

// Extracts all requires ( {{require example.t}} ) from a given file.
// Takes into account fallback files too.
// First it checks if the file exists in the current template. If yes, the link will point to that file.
// If not, then the link will point to the fallback module file.
// TODO: Case when the required file does not exists anywhere is not handled.
func ReqLinks(opt map[string]interface{}, file, root, host string) []ReqLink {
	pos := require.RequirePositions(file)
	ret := []ReqLink{}
	for _, v := range pos {
		fi := file[v[0]+10 : v[1]-2] // cut {{require anything/anything.t}} => anything/anything.t
		var typ, path, name string
		exists_in_template, err := Exists(filepath.Join(root, scut.GetTPath(opt, host), fi))
		if err != nil {
			continue
		}
		if exists_in_template {
			typ = scut.TemplateType(opt)
			path = fi
			name = scut.TemplateName(opt)
		} else {
			path = scut.GetModTPath(fi)[1]
			typ = "mod"
			name = strings.Split(fi, "/")[0]
		}
		rl := ReqLink{typ, name, path}
		ret = append(ret, rl)
	}
	return ret
}

type Breadc struct {
	Name string
	Path string
}

// fs is strings.Split(filepath, "/") where filepath is "aboutus/joe.tpl"
func CreateBreadCrumb(fs []string) []Breadc {
	ret := []Breadc{}
	for i := 1; i < len(fs); i++ {
		str := strings.Replace(filepath.Join(fs[:i+1]...), "\\", "/", -1)
		ret = append(ret, Breadc{fs[i], "/" + str})
	}
	return ret
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

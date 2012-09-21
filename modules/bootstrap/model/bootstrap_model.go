package bootstrap_model

import(
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"regexp"
	"time"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"sync"
	"os"
	"os/exec"
	"net"
	"math/rand"
	"github.com/opesun/jsonp"
	"github.com/opesun/extract"
	"github.com/opesun/numcon"
	"github.com/opesun/hypecms/modules/admin/model"
	"github.com/opesun/hypecms/model/basic"
	"strings"
)

type m map[string]interface{}

// Returns false if the server exceeded its capacity.
func hasRoom(db *mgo.Database, max_cap int) (bool, error) {
	count, err := db.C("sites").Find(nil).Count()
	if err != nil {
		return false, err
	}
	return count < max_cap, nil
}

// Saves a new site into the sites collection, along with the password of the database access.
// Saves data without any validation or whatsoever!
func newSite(db *mgo.Database, sitename, db_pass string) error {
	s := map[string]interface{}{
		"sitename": sitename,
		"db_password": db_pass,
	}
	return db.C("sites").Insert(s)
}

// Currently we wonly delete a site from the sites collection of the root database, for safety reasons.
// The deletion will only take effect at next restart.
func DeleteSite(db *mgo.Database, inp map[string][]string) error {
	r := map[string]interface{}{
		"sitename": "must",
	}
	dat, err := extract.New(r).Extract(inp)
	if err != nil {
		return err
	}
	return deleteSite(db, dat["sitename"].(string))
}

// Deletes a site.
func deleteSite(db *mgo.Database, sitename string) error {
	q := map[string]interface{}{
		"sitename": sitename,
	}
	return db.C("sites").Remove(q)
}

// Returns true if sitename is a valid subdomain.
func validSitename(sitename string) bool {
	re := regexp.MustCompile(`^[a-z\d]+(-[a-z\d]+)*$`)
	return re.MatchString(sitename)
}

// Returns false if there is already a site which is named sitename.
func sitenameAvailable(db *mgo.Database, sitename string) (bool, error) {
	var v []interface{}
	q := map[string]interface{}{"sitename": sitename}
	err := db.C("sites").Find(q).All(&v)
	if err != nil {
		return false, err
	}
	if len(v) == 0 {
		return true, nil
	}
	return false, nil
}

type SiteInfo struct {
	Name		string
	DbPassword	string
}

func AllSitenames(db *mgo.Database) ([]string, error) {
	sinfos, err := allSites(db)
	if err != nil {
		return nil, err
	}
	sitenames := []string{}
	for _, v := range sinfos {
		sitenames = append(sitenames, v.Name)
	}
	return sitenames, nil
}

func SiteCount(db *mgo.Database) (int, error) {
	return db.C("sites").Find(nil).Count()
}

// Returns all sitenames in database.
func allSites(db *mgo.Database) ([]SiteInfo, error) {
	var sites []interface{}
	err := db.C("sites").Find(nil).All(&sites)
	if err != nil {
		return nil, err
	}
	sinfos := []SiteInfo{}
	for _, v := range sites {
		sitename, has := jsonp.GetStr(v, "sitename")
		if !has {
			fmt.Println("Site has no name.")
			continue
		}
		db_password, has := jsonp.GetStr(v, "db_password")
		if !has {
			fmt.Println(fmt.Sprintf("Site %v has no db password.", sitename))
		}
		sinfo := SiteInfo{sitename, db_password}
		sinfos = append(sinfos, sinfo)
	}
	return sinfos, nil
}

// Returns true if it is okay to start a new site with arg sitename.
func sitenameIsOk(db *mgo.Database, sitename string) error {
	if !validSitename(sitename) {
		return fmt.Errorf("Sitename must be a valid subdomain.")
	}
	name_avail, err := sitenameAvailable(db, sitename)
	if err != nil {
		return err
	}
	if !name_avail {
		return fmt.Errorf("Sitename is already taken.")
	}
	return nil
}

var alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// generates a random string of fixed size
// Used in bootstrap to generate random database password.
func srand(size int) string {
    buf := make([]byte, size)
    for i := 0; i < size; i++ {
        buf[i] = alpha[rand.Intn(len(alpha)-1)]
    }
    return string(buf)
}

// Writes the entries in sitename_to_port into the proxy table.
//
// sitename_to_port example:
// {
// "example1": 51986
// }
//
// Can be used to write one or more (all, see bootstrap.StartAll) sites to the proxy table.
// table_key now does not support dot notation, eg "x.y." will not create x and y if they don't exist.
func writeProxyTable(proxy_abs, table_key, host_format string, sitename_to_port map[string]int) error {
	mut := new(sync.Mutex)
	mut.Lock()
	defer mut.Unlock()
	file_b, err := ioutil.ReadFile(proxy_abs)
	if err != nil {
		return err
	}
	var proxy_table_i interface{}
	err = json.Unmarshal(file_b, &proxy_table_i)
	if err != nil {
		return err
	}
	place, ok := jsonp.GetM(proxy_table_i, table_key)
	if !ok {
		return fmt.Errorf("Proxy table member %v is not a map.", table_key)
	}
	for i, v := range sitename_to_port {
		proxy_from := fmt.Sprintf(host_format, i)
		proxy_to := []interface{}{
			"http://127.0.0.1:" + strconv.Itoa(v),
		}
		place[proxy_from] = proxy_to
	}
	marshalled_table, err := json.MarshalIndent(proxy_table_i, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(proxy_abs, marshalled_table, os.ModePerm)
}

// Generate a random number in a range of probably free ports.
func randPort() int {
	return 53000 + rand.Intn(11000)
}

// Tries to find a free port.
// Birthday paradox will kick in pretty soon, that's why we retry max_tries number of times.
func generatePortnum() (int, error) {
	port_num := randPort()
	_, err := net.Dial("tcp", "127.0.0.1:" + strconv.Itoa(port_num))
	if err != nil {				// If a port is unused, dials gives you an error.
		return port_num, nil
	}
	return 0, fmt.Errorf("Can not find free port.")
}

// Gives you n free ports.
func freePorts(max_tries, n int) ([]int, error) {
	rand.Seed(time.Now().Unix())
	portnums := map[int]struct{}{}
	for i:=0;i<max_tries;i++{
		num, err := generatePortnum()
		if err == nil {
			portnums[num] = struct{}{}
		}
		if len(portnums) == n {
			break
		}
	}
	if len(portnums) < n {
		return nil, fmt.Errorf("Could not find %v free ports with %v tries.", n, max_tries)
	}
	ret := []int{}
	for i := range portnums {
		ret = append(ret, i)
	}
	return ret, nil
}

// Gives you a free port.
func freePort(max_tries int) (int, error) {
	free, err := freePorts(max_tries, 1)
	if err != nil {
		return 0, err
	}
	return free[0], nil
}

// Starts the executable for sitename.
func startExecInstance(sys_root, exec_abs, sitename, db_password string, port_num int) error {
	port_arg := 	"-p=" + strconv.Itoa(port_num)
	abs_path :=		"-abs_path=" + sys_root
	db_name_arg := 	"-db_name=" + sitename
	db_user_arg := 	"-db_user=" + sitename
	db_pass_arg := 	"-db_pass=" + db_password
	secret		:=	"-secret=" + db_password	// Think about this again.
	cmd := exec.Command(exec_abs, abs_path, port_arg, db_name_arg, db_user_arg, db_pass_arg, secret)
	return cmd.Start()
}

func newSiteRules() map[string]interface{} {
	subr := map[string]interface{}{
		"must": 1,
		"type": "string",
		"min": 2,
		"max": 30,
	}
	return map[string]interface{}{
		"sitename": subr,
		"password": subr,
		"password_again": subr,
	}
}

func defaultOpts(db *mgo.Database) (map[string]interface{}, error) {
	// Dohh, a different collection for a single doc, really unmongoish, rewrite.
	var res interface{}
	err := db.C("default_opt").Find(nil).One(&res)
	if err != nil {
		return nil, err
	}
	return basic.Convert(res).(map[string]interface{}), nil
}

func igniteReadOps(session *mgo.Session, db *mgo.Database, boots_opt map[string]interface{}, inp map[string][]string) (map[string]interface{}, string, error) {
	if session == nil {
		return nil, "", fmt.Errorf("This is not an admin instance.")
	}
	r := newSiteRules()
	dat, err := extract.New(r).Extract(inp)
	if err != nil {
		return nil, "", err
	}
	// Need to check it here too, not just at admin_model.RegFirstAdmin!
	if dat["password"].(string) != dat["password_again"].(string) {
		return nil, "", fmt.Errorf("Passwords do not match.")
	}
	max_cap := numcon.IntP(boots_opt["max_cap"])
	hasroom, err := hasRoom(db, max_cap)
	if err != nil {
		return nil, "", err
	}
	if !hasroom {
		return nil, "", fmt.Errorf("Server is at full capacity.")
	}
	sitename := dat["sitename"].(string)
	root_db := boots_opt["root_db"].(string)
	if sitename == root_db || strings.HasPrefix(sitename, "www") {
		return nil, "", fmt.Errorf("Sitename cant equal to root sitename or start with www.")
	}
	default_must := boots_opt["default_must"].(bool)
	def_opts, err := defaultOpts(db)
	if default_must && err != nil {
		return nil, "", fmt.Errorf("Cant read default option document.")
	}
	return def_opts, sitename, sitenameIsOk(db, sitename)
}

func saveDefaultOpt(site_db *mgo.Database, def_opts map[string]interface{}) error {
	if def_opts == nil {
		return fmt.Errorf("Nothing to save.")
	}
	id := basic.CreateOptCopy(site_db)
	return site_db.C("options").Update(m{"_id": id}, m{"$set": def_opts})
}

func igniteWriteOps(session *mgo.Session, db *mgo.Database, boots_opt map[string]interface{}, inp map[string][]string, def_opts map[string]interface{}, sitename string) error {
	db_password := srand(20)
	err := newSite(db, sitename, db_password)
	if err != nil {
		return err
	}
	site_db := session.DB(sitename)
	default_must := boots_opt["default_must"].(bool)
	err = saveDefaultOpt(site_db, def_opts)
	if default_must && err != nil {
		return err
	}
	err = site_db.AddUser(sitename, db_password, false)
	if err != nil {
		return err
	}
	err = admin_model.RegFirstAdmin(site_db, inp)
	if err != nil {
		return err
	}
	// Since everyone tried to login as sitename, we regin an admin here named sitename.
	inp["name"] = []string{sitename}
	err = admin_model.RegAdmin(site_db, inp)
	if err != nil {
		return err
	}
	port_num, err := freePort(10)
	if err != nil {
		return err
	}
	proxy_abs := boots_opt["proxy_abs"].(string)
	table_key := boots_opt["table_key"].(string)
	host_format := boots_opt["host_format"].(string)
	err = writeProxyTable(proxy_abs, table_key, host_format, map[string]int{sitename: port_num})
	if err != nil {
		return err
	}
	exec_abs := boots_opt["exec_abs"].(string)
	sys_root := boots_opt["sys_root"].(string)
	err = startExecInstance(sys_root, exec_abs, sitename, db_password, port_num)
	return err
}

func igniteCleanUp(session *mgo.Session, db *mgo.Database, sitename string) error {
	site_db := session.DB(sitename)
	names, err := site_db.CollectionNames()
	if err != nil {
		return err
	}
	fmt.Println(names)
	if len(names) > 4 {		// Freshly created database will have system.indexes, system.users, users and options collections at max.
		return fmt.Errorf("Can't delete an existing database with meaningful data.")
	}
	deleteSite(db, sitename)	// Ignore errors.
	return site_db.DropDatabase()
}

// This registers the site into the sites collection, creates a database for it, and registers an admin for it with the password coming
// from user interface. Also writes the site into the proxy table and starts the http server for the new site.
func Ignite(session *mgo.Session, db *mgo.Database, boots_opt map[string]interface{}, inp map[string][]string) (sitename string, err error) {
	mut := new(sync.Mutex)
	mut.Lock()
	defer mut.Unlock()
	def_opts, sitename, err := igniteReadOps(session, db, boots_opt, inp)
	if err != nil {
		return
	}
	defer func() {
		r := recover()
		if r != nil || err != nil {
			if r != nil {
				err = fmt.Errorf(fmt.Sprint(r))
			}
			err1 := igniteCleanUp(session, db, sitename)
			if err1 != nil {
				// What to do here?
				fmt.Println("Unable to clear up:", err1)
			}
		}
	}()
	err = igniteWriteOps(session, db, boots_opt, inp, def_opts, sitename)
	return
}

// Starts a process for each site in the sites collection.
func StartAll(db *mgo.Database, boots_opt map[string]interface{}) error {
	sinfos, err := allSites(db)
	if err != nil {
		return err
	}
	l := len(sinfos)
	if l == 0 {
		return fmt.Errorf("No sites to start.")		// Not really mean't as an error, more like information.
	}
	proxy_abs := boots_opt["proxy_abs"].(string)
	table_key := boots_opt["table_key"].(string)
	host_format := boots_opt["host_format"].(string)
	exec_abs := boots_opt["exec_abs"].(string)
	port_nums, err := freePorts(l*10, len(sinfos))
	if err != nil {
		return err
	}
	proxy_data := map[string]int{}
	for i, v := range sinfos {
		proxy_data[v.Name] = port_nums[i]
	}
	err = writeProxyTable(proxy_abs, table_key, host_format, proxy_data)
	if err != nil {
		return err
	}
	sys_root := boots_opt["sys_root"].(string)
	for i, v := range sinfos {
		err := startExecInstance(sys_root, exec_abs, v.Name, v.DbPassword, port_nums[i])
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to start %v.", v.Name))
		}
	}
	return nil
}

func Install(session *mgo.Session, db *mgo.Database, id bson.ObjectId) error {
	if session == nil {
		return fmt.Errorf("This is not an admin instance.")
	}
	bootstrap_options := m{
	}
	q := m{"_id": id}
	upd := m{
		"$addToSet": m{
			"Hooks.BeforeDisplay": "bootstrap",
		},
		"$set": m{
			"Modules.bootstrap": bootstrap_options,
		},
	}
	return db.C("options").Update(q, upd)
}

func Uninstall(db *mgo.Database, id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$pull": m{
			"Hooks.BeforeDisplay": "bootstrap",
		},
		"$unset": m{
			"Modules.bootstrap": 1,
		},
	}
	return db.C("options").Update(q, upd)
}
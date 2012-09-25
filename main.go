// hypeCMS is a CMS and/or framework for web applications, and more.
// For license, see the file named LICENSE.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/mod"
	"github.com/opesun/hypecms/api/shell"
	"github.com/opesun/hypecms/model/main"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/admin"
	"github.com/opesun/hypecms/modules/display"
	"github.com/opesun/hypecms/modules/user"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"strings"
)

const (
	unfortunate_error         = "main: An unfortunate error has happened. We are deeply sorry for the inconvenience."
	unexported_front          = "main: Module %v does not export Front view."
	unexported_action         = "main: Module %v does not export action %v."
	no_user_module_build_hook = "main: User module does not export BuildUser hook."
	no_module_at_action       = "main: Tried to execute action, but no module was specified."
	no_action                 = "main: No action specified when accessing module %v."
	no_admin_action		      = "main: No admin action specified."
)

// See handleFlags methods about these vars and their uses.
var (
	ABS_PATH    string
	CONF_FN     string
	DB_ADM_MODE bool
	DB_USER     string
	DB_PASS     string
	DB_ADDR     string
	DEBUG       bool
	DB_NAME     string
	ADDR        string
	PORT_NUM    string
	OPT_CACHE   bool
	SERVE_FILES bool
	SECRET      string
)

func loadConfFromFile() {
	cf, err := ioutil.ReadFile(filepath.Join(ABS_PATH, CONF_FN))
	if err != nil {
		fmt.Println("Could not read the config file, falling back to defaults.")
		return
	}
	var conf_i interface{}
	err = json.Unmarshal(cf, &conf_i)
	if err != nil || conf_i == nil {
		fmt.Println("Could not decode config json file, falling back to defaults.")
		return
	}
	conf, ok := conf_i.(map[string]interface{})
	if !ok {
		fmt.Println("Config is not a map, falling back to defaults.")
		return
	}
	if db_adm_mode, ok := conf["db_admin_mode"].(bool); ok {
		DB_ADM_MODE = db_adm_mode
	}
	// Doh...
	if db_user, ok := conf["db_user"].(string); ok {
		DB_USER = db_user
	}
	if db_pass, ok := conf["db_pass"].(string); ok {
		DB_PASS = db_pass
	}
	if db_addr, ok := conf["db_addr"].(string); ok {
		DB_ADDR = db_addr
	}
	if debug, ok := conf["debug"].(bool); ok {
		DEBUG = debug
	}
	if db_name, ok := conf["db_name"].(string); ok {
		DB_NAME = db_name
	}
	if addr, ok := conf["addr"].(string); ok {
		ADDR = addr
	}
	if port_num, ok := conf["port_num"].(string); ok {
		PORT_NUM = port_num
	}
	if opt_cache, ok := conf["opt_cache"].(bool); ok {
		OPT_CACHE = opt_cache
	}
	if serve_files, ok := conf["serve_files"].(bool); ok {
		SERVE_FILES = serve_files
	}
	if secret, ok := conf["secret"].(string); ok {
		SECRET = secret
	}
}

func handleConfigVars() {
	flag.StringVar(	&ABS_PATH, 		"abs_path", 	"c:/gowork/src/github.com/opesun/hypecms", "absolute path")
	flag.StringVar(	&CONF_FN, 		"conf_fn", 		"config.json", 		"config filename")
	// Everything else we can try to load from file.
	loadConfFromFile()
	flag.BoolVar(	&DB_ADM_MODE, 	"db_adm_mode", 	false, 				"connect to database as an admin")
	flag.StringVar(	&DB_USER, 		"db_user", 		"", 				"database username")
	flag.StringVar(	&DB_PASS, 		"db_pass", 		"", 				"database password")
	flag.StringVar(	&DB_ADDR, 		"db_addr", 		"127.0.0.1:27017", 	"database address")
	flag.BoolVar(	&DEBUG, 		"debug", 		true, 				"debug mode")
	flag.StringVar(	&DB_NAME, 		"db_name", 		"hypecms", 			"db name to connect to")
	flag.StringVar(	&PORT_NUM, 		"p", 			"80", 				"port to listen on")
	flag.StringVar(	&ADDR, 			"addr", 		"", 				"address to start http server")
	flag.BoolVar(	&OPT_CACHE, 	"opt_cache", 	false, 				"cache option document")
	flag.BoolVar(	&SERVE_FILES, 	"serve_files", 	true, 				"serve files from Go or not")
	flag.StringVar(	&SECRET, 		"secret", 		"pLsCh4nG3Th1$.AlSoThisShouldbeatLeast16bytes", "secret characters used for encryption and the like")
	flag.Parse()
}

// Quickly print the data to http response.
var Put func(...interface{})

type m map[string]interface{}

// All front hooks must have the signature of func(*context.Uni, *bool) error
// All views are going to use this hook.
func execFrontViews(uni *context.Uni) {
	var err error
	i := func(hijacked bool, er error) bool {
		if er != nil {
			err = er
			return true
		}
		return hijacked
	}
	uni.Ev.Iterate("Front", i)
	if err == nil {
		display.D(uni)
	} else {
		display.DErr(uni, err)
	}
}

// This writes all necessary information after a background operation into the redirect url, and deletes
// parts which were when a previous background op ran.
func appendParams(url_str string, action_name string, err error, cont map[string]interface{}) string {
	p := strings.Split(url_str, "?")
	var inp string
	if len(p) > 1 {
		inp = p[1]
	} else {
		inp = ""
	}
	v, parserr := url.ParseQuery(inp)
	if parserr == nil {
		// Delete outdated information from url.
		for i := range v {
			if strings.HasPrefix(i, "-") {
				v.Del(i)
			}
		}
		// Write all data in cont into the url.
		for key, val := range cont {
			if key[0] == '!' {
				v.Set(key[1:], fmt.Sprint(val))
			} else {
				v.Set("-"+key, fmt.Sprint(val))
			}
		}
		v.Del("error")
		v.Del("ok") // See *1
		v.Del("action")
		if len(action_name) > 0 { // runDebug calls this function with an empty action name.
			v.Set("action", action_name)
		}
		if err == nil {
			v.Set("ok", "true") // This could be left out, but hey. *1
		} else {
			v.Set("error", err.Error())
		}
		quer := v.Encode()
		if len(quer) > 0 {
			return p[0] + "?" + quer
		}
	}
	return p[0]
}

// After running a background operation this either redirects with data in url paramters or prints out the json encoded result.
func actionResponse(uni *context.Uni, err error, action_name string) {
	if DEBUG {
		fmt.Println(uni.Req.Referer())
		fmt.Println("	", err)
	}
	_, is_json := uni.Req.Form["json"]
	redir := uni.Req.Referer()
	if red, ok := uni.Dat["redirect"]; ok {
		redir = red.(string)
	} else if post_red, okr := uni.Req.Form["redirect"]; okr && len(post_red) == 1 {
		redir = post_red[1]
	}
	var cont map[string]interface{}
	cont_i, has := uni.Dat["_cont"]
	if has {
		cont = cont_i.(map[string]interface{})
	} else {
		cont = map[string]interface{}{}
	}
	redir = appendParams(redir, action_name, err, cont)
	if is_json {
		cont["redirect"] = redir
		if err == nil {
			cont["ok"] = true
		} else {
			cont["error"] = err.Error()
		}
		var v []byte
		if _, fmt := uni.Req.Form["fmt"]; fmt {
			v, _ = json.MarshalIndent(cont, "", "    ")
		} else {
			v, _ = json.Marshal(cont)
		}
		uni.Put(string(v))
	} else {
		http.Redirect(uni.W, uni.Req, redir, 303)
	}
}

func sanitizeActionname(a string) string {
	a = strings.Replace(a, "-", " ", -1)
	a = strings.Replace(a, "_", " ", -1)
	a = strings.Title(a)
	return strings.Replace(a, " ", "", -1)
}

// All back hooks must have the signature of func(*context.Uni, string) error
func runAction(uni *context.Uni) (string, error) {
	l := len(uni.Paths)
	if l < 3 {
		return "", fmt.Errorf(no_module_at_action)
	}
	modname := uni.Paths[2] // TODO: Routing based on Paths won't work if the site is installed to subfolder or something.
	if l < 4 {
		return "", fmt.Errorf(no_action, modname)
	}
	action_name := uni.Paths[3]
	err, puzzle_err := user.OkayToDoAction(uni, modname, action_name)
	if err != nil {
		return action_name, err
	}
	if puzzle_err != nil {
		return action_name, puzzle_err
	}
	sanitized_aname := sanitizeActionname(action_name)
	if !uni.Caller.Has("actions", modname, sanitized_aname) {
		return action_name, fmt.Errorf(unexported_action, modname)
	}
	if !uni.Caller.Matches("actions", modname, sanitized_aname, func() error {return nil}) {
		return action_name, fmt.Errorf("Action %v of %v has bad signature.", action_name, modname)
	}
	ret_rec := func(e error){
		err = e
	}
	uni.Caller.Call("actions", modname, sanitized_aname, ret_rec)
	return action_name, err
}

// Every background operation uses this hook.
func execAction(uni *context.Uni) {
	action_name, err := runAction(uni)
	actionResponse(uni, err, action_name)
}

func execAdmin(uni *context.Uni) {
	l := len(uni.Paths)
	var err error
	if l > 2 && uni.Paths[2] == "b" {
		var action_name string
		if l > 3 {
			action_name = uni.Paths[3]
			uni.Dat["_action"] = action_name
			err = admin.AB(uni, action_name)
		} else {
			err = fmt.Errorf(no_admin_action)
		}
		actionResponse(uni, err, action_name)
	} else {
		err = admin.AD(uni)
		if err == nil {
			display.D(uni)
		} else {
			display.DErr(uni, err)
		}
	}
}

func buildUser(uni *context.Uni) error {
	// Why is this a hook? Get rid of it.
	if !uni.Caller.Has("hooks", "user", "BuildUser") {
		return fmt.Errorf(no_user_module_build_hook)
	}
	var err error
	ret_rec := func(e error) {
		e = err
	}
	uni.Caller.Call("hooks", "user", "BuildUser", ret_rec)
	return err
}

func terminal(uni *context.Uni) {
	shell.Terminal(uni)
	display.D(uni)
}

func execCommands(uni *context.Uni) {
	err := shell.FromWeb(uni)
	actionResponse(uni, err, "shell")
}

func runSite(uni *context.Uni) {
	err := buildUser(uni)
	if err != nil {
		display.DErr(uni, err)
		return
	}
	switch uni.Paths[1] {
	// Back hooks are put behind "/b/" to avoid eating up the namespace.
	case "b":
		execAction(uni)
	// Admin is a VIP module, to allow bootstrapping a site even if the option document is empty.
	case "admin":
		execAdmin(uni)
	case "run-commands":
		execCommands(uni)
	case "terminal":
		terminal(uni)
	default:
		execFrontViews(uni)
	}
}

// Just printing the stack trace to http response if a panic bubbles up all the way to top.
func err() {
	if r := recover(); r != nil {
		fmt.Println("at main:", r)
		fmt.Println(string(debug.Stack()))
		Put(unfortunate_error)
		Put(fmt.Sprint("\n", r, "\n\n"+string(debug.Stack())))
	}
}

// getSite gets the freshest option document, caches it and creates an instance of context.Uni.
func getSite(session *mgo.Session, db *mgo.Database, w http.ResponseWriter, req *http.Request) {
	Put = func(a ...interface{}) {
		io.WriteString(w, fmt.Sprint(a...)+"\n")
	}
	defer err()
	uni := &context.Uni{
		Db:      db,
		W:       w,
		Req:     req,
		Put:     Put,
		Dat:     make(map[string]interface{}),
		Root:    ABS_PATH,
		P:       req.URL.Path,
		Paths:   strings.Split(req.URL.Path, "/"),
	}
	uni.Caller = mod.NewCall(uni)
	// Not sure if not giving the db session to nonadmin installations increases security, but hey, one can never be too cautious, they dont need it anyway.
	if DB_ADM_MODE {
		uni.Session = session
	}
	uni.Ev = context.NewEv(uni)
	opt, opt_str, err := main_model.HandleConfig(uni.Db, req.Host, OPT_CACHE) // Tricky part about the host, see comments at main_model.
	if err != nil {
		uni.Put(err.Error())
		return
	}
	uni.Req.Host = scut.Host(req.Host, opt)
	uni.Opt = opt
	uni.SetOriginalOpt(opt_str)
	uni.SetSecret(SECRET)
	first_p := uni.Paths[1]
	last_p := uni.Paths[len(uni.Paths)-1]
	if SERVE_FILES && strings.Index(last_p, ".") != -1 {
		has_sfx := strings.HasSuffix(last_p, ".go")
		if first_p == "template" || first_p == "tpl" && !has_sfx {
			serveTemplateFile(w, req, uni)
		} else if !has_sfx {
			if uni.Paths[1] == "shared" {
				http.ServeFile(w, req, filepath.Join(ABS_PATH, req.URL.Path))
			} else {
				http.ServeFile(w, req, filepath.Join(ABS_PATH, "uploads", req.Host, req.URL.Path))
			}
		} else {
			uni.Put("Don't do that.")
		}
		return
	}
	req.ParseForm()		// Should we handle the error return of this?
	runSite(uni)
}

// Since we don't include the template name into the url, only "template", we have to extract the template name from the opt here.
// Example: xyz.com/template/style.css
//			xyz.com/tpl/admin/style.css
func serveTemplateFile(w http.ResponseWriter, req *http.Request, uni *context.Uni) {
	if uni.Paths[1] == "template" {
		p := scut.GetTPath(uni.Opt, uni.Req.Host)
		http.ServeFile(w, req, filepath.Join(uni.Root, p, strings.Join(uni.Paths[2:], "/")))
	} else { // "tpl"
		http.ServeFile(w, req, filepath.Join(uni.Root, "modules", uni.Paths[2], "tpl", strings.Join(uni.Paths[3:], "/")))
	}
}

func main() {
	fmt.Println("Server has started.")
	handleConfigVars()
	//if DEBUG {
	//	modcheck.Check()
	//}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	dial := DB_ADDR
	if len(DB_USER) != 0 || len(DB_PASS) != 0 {
		if len(DB_USER) == 0 {
			panic("Database password provided but username is missing.")
		}
		if len(DB_PASS) == 0 {
			panic("Database username is provided but password is missing.")
		}
		dial = DB_USER + ":" + DB_PASS + "@" + dial
		if !DB_ADM_MODE {
			dial = dial + "/" + DB_NAME
		}
	}
	session, err := mgo.Dial(dial)
	if err != nil {
		panic(err)
	}
	db := session.DB(DB_NAME)
	defer session.Close()
	http.HandleFunc("/",
	func(w http.ResponseWriter, req *http.Request) {
		getSite(session, db, w, req)
	})
	err = http.ListenAndServe(ADDR+":"+PORT_NUM, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("There were a problem when starting the http server.")
}

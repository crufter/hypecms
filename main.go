// HypeCMS is a CMS and/or framework for web applications, and more.
// Copyright Opesun Technologies Kft. 2012. See license.txt.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/mod"
	"github.com/opesun/hypecms/api/modcheck"
	"github.com/opesun/hypecms/model/main"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/admin"
	"github.com/opesun/hypecms/modules/display"
	"github.com/opesun/jsonp"
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
	unfortunate_error         = "An unfortunate error has happened. We are deeply sorry for the inconvenience."
	unexported_front          = "Module %v does not export Front hook."
	unexported_back           = "Module %v does not export Back hook."
	no_user_module_build_hook = "User module does not export build hook."
	no_module_at_back         = "Tried to run a back hook, but no module was specified."
	no_action                 = "No action specified when accessing module %v."
	adminback_no_module       = "No module specified when accessing admin back."
	cant_run_back             = "Can't run back hook of not installed module %v."
	cant_test                 = "Can't test module because it is not even installed: %v."
)

// See handleFlags methods about these vars and their uses.
var (
	ABS_PATH    string
	CONF_FN     string
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
		fmt.Println("Could not read the config file.")
		return
	}
	var conf_i interface{}
	err = json.Unmarshal(cf, &conf_i)
	if err != nil || conf_i == nil {
		fmt.Println("Could not decode config json file.")
		return
	}
	conf, ok := conf_i.(map[string]interface{})
	if !ok {
		fmt.Println("Config is not a map.")
		return
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
	flag.StringVar(&ABS_PATH, "abs_path", "c:/gowork/src/github.com/opesun/hypecms", "absolute path")
	flag.StringVar(&CONF_FN, "conf_fn", "config.json", "config filename")
	// Everything else we can try to load from file.
	loadConfFromFile()
	flag.StringVar(&DB_USER, "db_user", "", "database username")
	flag.StringVar(&DB_PASS, "db_pass", "", "database password")
	flag.StringVar(&DB_ADDR, "db_addr", "127.0.0.1:27017", "database address")
	flag.BoolVar(&DEBUG, "debug", true, "debug mode")
	flag.StringVar(&DB_NAME, "db_name", "hypecms", "db name to connect to")
	flag.StringVar(&PORT_NUM, "p", "80", "port to listen on")
	flag.StringVar(&ADDR, "addr", "", "address to start http server")
	flag.BoolVar(&OPT_CACHE, "opt_cache", false, "cache option document")
	flag.BoolVar(&SERVE_FILES, "serve_files", true, "serve files from Go or not")
	flag.StringVar(&SECRET, "secret", "pLsCh4nG3Th1$.AlSoThisShouldbeatLeast16bytes", "secret characters used for encryption and the like")
	flag.Parse()
}

// Quickly print the data to http response.
var Put func(...interface{})

type m map[string]interface{}

// All front hooks must have the signature of func(*context.Uni, *bool) error
// All views are going to use this hook.
func runFrontHooks(uni *context.Uni) {
	var err error
	top_hooks, ok := jsonp.GetS(uni.Opt, "Hooks.Front")
	if ok && len(top_hooks) > 0 {
		for _, v := range top_hooks {
			modname := v.(string)
			hijacked := false
			if h := mod.GetHook(modname, "Front"); h != nil {
				hook, ok := h.(func(*context.Uni, *bool) error)
				if !ok {
					err = fmt.Errorf("Front hook of %v has bad signature.", modname)
					break
				}
				err = hook(uni, &hijacked)
			} else {
				err = fmt.Errorf(unexported_front, modname)
				break
			}
			if hijacked {
				break
			}
		}
	}
	if err == nil {
		display.D(uni)
	} else {
		display.DErr(uni, err)
	}
}

// This is real basic yet, it would be cool to include all elements of result.
func appendParams(str string, err error, action_name string, cont map[string]interface{}) string {
	p := strings.Split(str, "?")
	var inp string
	if len(p) > 1 {
		inp = p[1]
	} else {
		inp = ""
	}
	v, parserr := url.ParseQuery(inp)
	if parserr == nil {
		for key, val := range cont { // This way we can include additional data in the get params, not only action name and errors.
			v.Set(key, fmt.Sprint(val))
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
func handleBacks(uni *context.Uni, err error, action_name string) {
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
	if is_json {
		cont["redirect"] = redir
		var v []byte
		if _, fmt := uni.Req.Form["fmt"]; fmt {
			v, _ = json.MarshalIndent(cont, "", "    ")
		} else {
			v, _ = json.Marshal(cont)
		}
		uni.Put(string(v))
	} else {
		redir = appendParams(redir, err, action_name, cont)
		http.Redirect(uni.W, uni.Req, redir, 303)
	}
}

// All back hooks must have the signature of func(*context.Uni, string) error
func runBacks(uni *context.Uni) (string, error) {
	l := len(uni.Paths)
	if l < 3 {
		return "", fmt.Errorf(no_module_at_back)
	}
	modname := uni.Paths[2] // TODO: Routing based on Paths won't work if the site is installed to subfolder or something.
	if l < 4 {
		return "", fmt.Errorf(no_action, modname)
	}
	action_name := uni.Paths[3]
	if _, installed := jsonp.Get(uni.Opt, "Modules."+modname); !installed {
		return action_name, fmt.Errorf(cant_run_back, modname)
	}
	h := mod.GetHook(modname, "Back")
	if h == nil {
		return action_name, fmt.Errorf(unexported_back, modname)
	}
	hook, ok := h.(func(*context.Uni, string) error)
	if !ok {
		return action_name, fmt.Errorf("Back hooks of %v has bad signature.", modname)
	}
	err := hook(uni, action_name)
	return action_name, err
}

// Every background operation uses this hook.
func runBackHooks(uni *context.Uni) {
	action_name, err := runBacks(uni)
	handleBacks(uni, err, action_name)
}

func runAdminHooks(uni *context.Uni) {
	l := len(uni.Paths)
	var err error
	if l > 2 && uni.Paths[2] == "b" {
		var action_name string
		if l > 3 {
			action_name = uni.Paths[3]
			uni.Dat["_action"] = action_name
			err = admin.AB(uni, action_name)
		} else {
			err = fmt.Errorf(adminback_no_module)
		}
		handleBacks(uni, err, action_name)
	} else {
		err = admin.AD(uni)
		if err == nil {
			display.D(uni)
		} else {
			display.DErr(uni, err)
		}
	}
}

func runD(uni *context.Uni) error {
	if len(uni.Paths) < 3 {
		return fmt.Errorf("No module specified to test.")
	}
	modname := uni.Paths[2]
	if _, installed := jsonp.Get(uni.Opt, "Modules."+modname); !installed {
		return fmt.Errorf(cant_test, modname)
	}
	h := mod.GetHook(modname, "Test")
	if h == nil {
		return fmt.Errorf("Module %v does not export Test hook.", modname)
	}
	hook, ok := h.(func(*context.Uni) error)
	if !ok {
		return fmt.Errorf("Test hook of %v has bad signature.", modname)
	}
	return hook(uni)
}

// Usage: /debug/{modulename} runs the test of the given module which compares the current option document to the "standard one" expected by the given module.
func runDebug(uni *context.Uni) {
	err := runD(uni)
	handleBacks(uni, err, "debug")
}

func buildUser(uni *context.Uni) error {
	h := mod.GetHook("user", "BuildUser")
	if h != nil {
		return h.(func(*context.Uni) error)(uni)
	}
	return fmt.Errorf(no_user_module_build_hook)
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
		runBackHooks(uni)
	// Admin is a VIP module, to allow bootstrapping a site even if the option document is empty.
	case "admin":
		runAdminHooks(uni)
	// Debug is VIP to allow debugging even with a messed up option document.
	case "debug":
		runDebug(uni)
	default:
		runFrontHooks(uni)
	}
}

// Just printing the stack trace to http response if a panic bubbles up all the way to top.
func err() {
	if r := recover(); r != nil {
		fmt.Println(r)
		Put(unfortunate_error)
		Put(fmt.Sprint("\n", r, "\n\n"+string(debug.Stack())))
	}
}

// getSite gets the freshest option document, caches it and creates an instance of context.Uni.
func getSite(db *mgo.Database, w http.ResponseWriter, req *http.Request) {
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
		GetHook: mod.GetHook,
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
	req.ParseForm()
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
	handleConfigVars()
	if DEBUG {
		modcheck.Check()
	}
	fmt.Println("Server has started.")
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
	}
	session, err := mgo.Dial(DB_ADDR)
	if err != nil {
		panic(err)
	}
	db := session.DB(DB_NAME)
	defer session.Close()
	http.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			getSite(db, w, req)
		})
	http.ListenAndServe(ADDR+":"+PORT_NUM, nil)
}

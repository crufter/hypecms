// HypeCMS is a CMS and/or framework for web applications, and more.
// Copyright Opesun Technologies Kft. 2012. See license.txt.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/mod"
	"github.com/opesun/hypecms/model/main"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/admin"
	"github.com/opesun/hypecms/modules/display"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"strings"
)

const (
	unfortunate_error			= "An unfortunate error has happened. We are deeply sorry for the inconvenience."
	unexported_front			= " module does not export Front hook."
	unexported_back				= " module does not export Back hook."
	no_user_module_build_hook	= "user module does not export build hook"
	no_module_at_back			= "Tried to run a back hook, but no module was specified."
	no_action					= "No action specified when accessing module "
	adminback_no_module			= "No module specified when accessing admin back."
	cant_run_back				= "Can't run back hook of not installed module "
	cant_test					= "Can't test module because it is not even installed: "
)

var DB_USER = ""
var DB_PASS = ""
var DB_ADDR = "127.0.0.1:27017"
var DEBUG = *flag.Bool("debug", true, "debug mode")
var DB_NAME = *flag.String("db", "hypecms", "db name to connect to")
var PORT_NUM = *flag.String("p", "80", "port to listen on")
var ABSOLUTE_PATH = "c:/gowork/src/github.com/opesun/hypecms"
var OPT_CACHE = *flag.Bool("opt_cache", false, "cache option document")
var SERVE_FILES = *flag.Bool("serve_files", true, "serve files from Go or not")

// Quickly print the data to http response.
var Put func(...interface{})

type m map[string]interface{}

// All views are going to use this hook.
func runFrontHooks(uni *context.Uni) {
	var err error
	top_hooks, ok := jsonp.GetS(uni.Opt, "Hooks.Front")
	if ok && len(top_hooks) > 0 {
		for _, v := range top_hooks {
			modname := v.(string)
			if h := mod.GetHook(modname, "Front"); h != nil {
				err = h(uni)
			} else {
				err = fmt.Errorf(modname + unexported_front)
				break
			}
			if _, ok := uni.Dat["_hijacked"]; ok {
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
func appendParams(str string, err error) string {
	p := strings.Split(str, "?")
	var inp string
	if len(p) > 1 {
		inp = p[1]
	} else {
		inp = ""
	}
	v, parserr := url.ParseQuery(inp)
	if parserr == nil {
		v.Del("error")
		v.Del("ok")
		if err == nil {
			v.Set("ok", "true")
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
func handleBacks(uni *context.Uni, err error) {
	if DEBUG {
		fmt.Println(uni.Req.Referer())
		fmt.Println("	", err)
	}
	_, is_json := uni.Req.Form["json"]
	if is_json {
		var v []byte
		if _, fmt := uni.Req.Form["fmt"]; fmt {
			v, _ = json.MarshalIndent(uni.Dat["_cont"], "", "    ")
		} else {
			v, _ = json.Marshal(uni.Dat["_cont"])
		}
		uni.Put(string(v))
	} else {
		redir := uni.Req.Referer()
		if red, ok := uni.Dat["redirect"]; ok {
			redir = red.(string)
		} else if post_red, okr := uni.Req.Form["redirect"]; okr && len(post_red) == 1 {
			redir = post_red[1]
		}
		redir = appendParams(redir, err)
		http.Redirect(uni.W, uni.Req, redir, 303)
	}
}

// Every background operation uses this hook.
func runBackHooks(uni *context.Uni) {
	var err error
	if len(uni.Paths) > 2 {
		modname := uni.Paths[2] // TODO: Routing based on Paths won't work if the site is installed to subfolder or something.
		if _, installed := jsonp.Get(uni.Opt, "Modules." + modname); !installed {
			err = fmt.Errorf(cant_run_back + modname)
		} else {
			if h := mod.GetHook(modname, "Back"); h != nil {
				if len(uni.Paths) > 3 {
					uni.Dat["_action"] = uni.Paths[3]
					err = h(uni)
				} else {
					err = fmt.Errorf(no_action + modname)
				}
			} else {
				err = fmt.Errorf(modname + unexported_back)
			}
		}
	} else {
		err = fmt.Errorf(no_module_at_back)
	}
	handleBacks(uni, err)
}

func runAdminHooks(uni *context.Uni) {
	l := len(uni.Paths)
	var err error
	if l > 2 && uni.Paths[2] == "b" {
		if l > 3 {
			uni.Dat["_action"] = uni.Paths[3]
			err = admin.AB(uni)
		} else {
			err = fmt.Errorf(adminback_no_module)
		}
		handleBacks(uni, err)
	} else {
		err = admin.AD(uni)
		if err == nil {
			display.D(uni)
		} else {
			display.DErr(uni, err)
		}
	}
}

// Usage: /debug/{modulename} runs the test of the given module which compares the current option document to the "standard one" expected by the given module.
func runDebug(uni *context.Uni) {
	var err error
	if len(uni.Paths) > 2 {
		modname := uni.Paths[2]
		if _, installed := jsonp.Get(uni.Opt, "Modules." + modname); !installed {
			err = fmt.Errorf(cant_test + modname)
		} else {
			err = mod.GetHook(modname, "Test")(uni)
		}
	} else {
		err = fmt.Errorf("No module specified to test.")
	}
	handleBacks(uni, err)
}

func buildUser(uni *context.Uni) error {
	h := mod.GetHook("user", "BuildUser")
	if h != nil {
		return h(uni)
	}
	return fmt.Errorf(no_user_module_build_hook)
}

// A runSite-ban van egy két hardcore-olt dolog (lásd forrást)
func runSite(uni *context.Uni) {
	buildUser(uni)
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

// A getSite gets the freshest option document, caches it and creates an instance of context.Uni.
func getSite(db *mgo.Database, w http.ResponseWriter, req *http.Request) {
	Put = func(a ...interface{}) {
		io.WriteString(w, fmt.Sprint(a...)+"\n")
	}
	defer err()
	uni := &context.Uni{
		Db:    		db,
		W:     		w,
		Req:   		req,
		Put:   		Put,
		Dat:   		make(map[string]interface{}),
		Root:  		ABSOLUTE_PATH,
		P:     		req.URL.Path,
		Paths: 		strings.Split(req.URL.Path, "/"),
		GetHook:	mod.GetHook,
	}
	uni.Ev = context.NewEv(uni)
	opt, err := main_model.HandleConfig(uni.Db, req.Host, OPT_CACHE)
	if err != nil {
		uni.Put(err.Error())
		return
	}
	uni.Opt = opt
	first_p := uni.Paths[1]
	last_p := uni.Paths[len(uni.Paths)-1]
	if SERVE_FILES && strings.Index(last_p, ".") != -1 {
		has_sfx := strings.HasSuffix(last_p, ".go")
		if first_p == "template" || first_p == "tpl" && !has_sfx {
			serveTemplateFile(w, req, uni)
		} else if !has_sfx {
			if uni.Paths[1] == "shared" {
				http.ServeFile(w, req, filepath.Join(ABSOLUTE_PATH, req.URL.Path))
			} else {
				http.ServeFile(w, req, filepath.Join(ABSOLUTE_PATH, "uploads", req.Host, req.URL.Path))
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
	} else {	// "tpl"
		http.ServeFile(w, req, filepath.Join(uni.Root, "modules", uni.Paths[2], "tpl", strings.Join(uni.Paths[3:], "/")))
	}
}

func main() {
	flag.Parse()
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
	db := session.DB(DB_NAME)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	http.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			getSite(db, w, req)
		})
	http.ListenAndServe("127.0.0.1:"+PORT_NUM, nil)
}

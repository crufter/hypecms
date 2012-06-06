// HypeCMS is a CMS and/or framework for webb applications, and more.
// Copyright OPESUN TECHNOLOGIES Kft. 2012. Authors: Donronszki János, Kapusi Lajos
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/mod"
	"github.com/opesun/hypecms/modules/admin"
	"github.com/opesun/hypecms/modules/display"
	"github.com/opesun/jsonp"
	"io"
	"launchpad.net/mgo"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
)

const (
	unfortunate_error 			= "An unfortunate error has happened. we are deeply sorry for the inconvenience."
	inv_userspace     			= "Userspace options string is not a valid JSON"	// TODO: Maybe we should try to recover from here.
	userspace_not_set			= "Userspace options are not set at all."
	unexported_front          	= " module does not export Front hook."
	unexported_back          	= " module does not export Back hook."
	no_user_module_build_hook 	= "user module does not export build hook"
	cant_encode_config        	= "Can't encode config. - No way this should happen anyway."
	no_module_at_back			= "Tried to run a back hook, but no module was specified."
)

var DB_ADDR = "127.0.0.1:27017"
var DEBUG = *flag.Bool("debug", true, "debug mode")
var DB_NAME = *flag.String("db", "hypecms", "db name to connect to")
var PORT_NUM = *flag.String("p", "80", "port to listen on")
var ABSOLUTE_PATH = "c:/gowork/src/github.com/opesun/hypecms"
var OPT_CACHE = *flag.Bool("opt_cache", false, "cache option document")

// Quickly print the data to http response.
var Put func(...interface{})

type m map[string]interface{}

// All views are going to use this hook.
func runFrontHooks(uni *context.Uni) {
	top_hooks, ok := jsonp.GetS(uni.Opt, "Hooks.Front")
	if ok && len(top_hooks) > 0 {
		for _, v := range top_hooks {
			modname := v.(string)
			if h := mod.GetHook(modname, "Front"); h != nil {
				h(uni)
			} else {
				Put(modname + unexported_front)
				return
			}
			if _, ok := uni.Dat["_hijacked"]; ok {
				break
			}
		}
	}
	display.D(uni)
}

// This is real basic yet, it would be cool to include all elements of result.
func appendParams(str string, result map[string]interface{}) string {
	p := strings.Split(str, "?")
	var inp string
	if len(p) > 1 {
		inp = p[1]
	} else {
		inp = ""
	}
	v, parserr := url.ParseQuery(inp)
	if parserr == nil {
		succ, ok := result["success"]
		if ok {
			v.Del("reason")		// When you do an illegal operation, result will be success=false,reason=x, but when you do a legal
			if succ == true {	// operation, you will be redirected to the same page. Result will be success=true,reason=x, to avoid it we do v.Del("reason")
				v.Set("success", "true")
			} else {
				v.Set("success", "false")
				reason, has_r := result["reason"]
				if has_r {
					v.Set("reason", reason.(string))
				}
			}
		}
		quer := v.Encode()
		fmt.Println(quer, len(quer))
		if len(quer) > 0 {
			return p[0] + "?" + quer
		}
	}
	return p[0]
}

// After running a background operation this either redirects with data in url paramters or prints out the json encoded result.
func handleBacks(uni *context.Uni) {
	if DEBUG {
		fmt.Println(uni.Req.Referer())
		fmt.Println("	", uni.Dat["_cont"])
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
		redir = appendParams(redir, uni.Dat["_cont"].(map[string]interface{}))
		http.Redirect(uni.W, uni.Req, redir, 303)
	}
}

// Every background operation uses this hook.
func runBackHooks(uni *context.Uni) {
	if len(uni.Paths) > 2 {
		modname := uni.Paths[2]		// TODO: Routing based on Paths won't work if the site is installed to subfolder or something.
		if h := mod.GetHook(modname, "Back"); h != nil {
			h(uni)
		} else {
			Put(modname + unexported_back)
			return
		}
		if _, ok := uni.Dat["_hijacked"]; ok {
			handleBacks(uni)
			return
		}
	} else {
		Put(no_module_at_back)
	}
}

func runAdminHooks(uni *context.Uni) {
	if len(uni.Paths) > 2 && uni.Paths[2] == "b" {
		admin.AB(uni)
		handleBacks(uni)
	} else {
		admin.AD(uni)
		display.D(uni)
	}
}

// Usage: /debug/{modulename} runs the test of the given module which compares the current option document to the "standard one" expected by the given module.
func runDebug(uni *context.Uni) {
	mod.GetHook(uni.Paths[2], "Test")(uni)
	handleBacks(uni)
}

func buildUser(uni *context.Uni) {
	h := mod.GetHook("user", "BuildUser")
	if h != nil {
		h(uni)
	} else {
		Put(no_user_module_build_hook)
		return
	}
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

func set(c map[string]string, key, val string) {
	mut := new(sync.Mutex)
	mut.Lock()
	c[key] = val
	mut.Unlock()
}

// mutex locked map get
func has(c map[string]string, str string) (interface{}, bool) {
	mut := new(sync.Mutex)
	mut.Lock()
	v, ok := c[str]
	mut.Unlock()
	return v, ok
}

// Just printing the stack trace to http response if a panic bubbles up all the way to top.
func err() {
	if r := recover(); r != nil && r != "controlled" {
		fmt.Println(r)
		Put(unfortunate_error)
		Put(fmt.Sprint("\n", r, "\n\n"+string(debug.Stack())))
	} else if r != nil && r == "controlled" {
		fmt.Println(r)
		Put(unfortunate_error)
	}
}

var cache = make(map[string]string)

// A getSite gets the freshest option document, caches it and creates an instance of context.Uni.
func getSite(db *mgo.Database, w http.ResponseWriter, req *http.Request) {
	Put = func(a ...interface{}) {
		io.WriteString(w, fmt.Sprint(a...) + "\n")
	}
	defer err()
	host := req.Host
	uni := &context.Uni{
		Db:   db,
		W:    w,
		Req:  req,
		Put:  Put,
		Dat:  make(map[string]interface{}),
		Root: ABSOLUTE_PATH,
		P:	  req.URL.Path,
		Paths: strings.Split(req.URL.Path, "/"),
	}
	if val, ok := has(cache, host); OPT_CACHE && ok {
		var v interface{}
		json.Unmarshal([]byte(val.(string)), &v)
		if v == nil {
			Put(inv_userspace)
			return
		}
		uni.Opt = v.(map[string]interface{})
		delete(uni.Opt, "_id")
	} else {
		var res interface{}
		db.C("options").Find(nil).Sort(m{"created": -1}).Limit(1).One(&res)
		if res == nil {
			res = m{}
			db.C("options").Insert(res)
		}
		enc, merr := json.Marshal(res)
		if merr != nil {
			Put(cant_encode_config)
			return
		}
		str := string(enc)
		set(cache, host, str)
		var v interface{}
		json.Unmarshal([]byte(str), &v)
		if v == nil {
			Put(inv_userspace)
			return
		}
		uni.Opt = v.(map[string]interface{})
		delete(uni.Opt, "_id")
	}
	req.ParseForm()
	runSite(uni)
}

func main() {
	flag.Parse()
	fmt.Println("server has started")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
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

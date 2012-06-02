// Tudod mi a fasz ez, így van?
package main

import(
	"net/http"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"io"
	"fmt"
	"sync"
	"encoding/json"
	"github.com/opesun/jsonp"
	"github.com/opesun/hypecms/api/mod"
	"github.com/opesun/hypecms/modules/admin"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/display"
	"runtime/debug"
	"strings"
	"flag"
)

const(
	unfortunate_error			= "an unfortunate error has happened. we are deeply sorry for the inconvenience."
	inv_userspace 				= "Userspace options string is not a valid JSON"
	site_not_found				= "site can not be found"
	userspace_not_set 			= "Userspace options are not set at all"
	front_hook_not_set			= "front hooks are not set properly"
	back_hook_not_set			= "back hooks are not set properly (either unset or empy slice)"
	unexported_front			= " module does not export Front hook"
	unexported_back				= "module's Back hook has bad signature"
	no_user_module_build_hook	= "user module does not export build hook"
	no_back_hijacked			= "none of the back hooks hijacked control"
)

var DB_ADDR = "127.0.0.1:27017"
var DEBUG = *flag.Bool("debug", true, "debug mode")
var DB_NAME = *flag.String("db", "hypecms", "db name to connect to")
var PORT_NUM = *flag.String("p", "80", "port to listen on")
var ABSOLUTE_PATH = "c:/gowork/src/github.com/opesun/hypecms"
// Http válaszban képernyőre írja az paramétereket.
var Put func(...interface{})
type m map[string]interface{}
	
	// A front hook egy olyan beépülési pont, ami látható kimenetet (viewt, azaz nézetet) fog produkálni. A runFronHooks az ide beépülő függvényeket futtatja.
	func runFrontHooks(uni *context.Uni) {
		top_hooks, ok := jsonp.GetS(uni.Opt, "Hooks.Front")
		if ok && len(top_hooks) > 0 {
			for _, v := range top_hooks {
				vs := v.(string)
				if h := mod.GetHook(vs, "Front"); h != nil {
					h(uni)
				} else {
					Put(vs + unexported_front)
					return
				}
				if _, ok := uni.Dat["_hijacked"]; ok {
					display.D(uni)
					return
				}
			}
			display.D(uni)
		} else {
			Put(front_hook_not_set)
		}
	}
		
		// A back hook lefutása után a handleBacks intézi vagy a JSON képernyőre írását, vagy a http redirectelést.
		func handleBacks(uni *context.Uni) {
			if DEBUG {
				fmt.Println(uni.Req.Referer())
				fmt.Println("	", uni.Dat["_cont"])
			}
			_, is_json := uni.Req.Form["json"]
			if is_json {
				v, _ := json.Marshal(uni.Dat["_cont"])
				uni.Put(string(v))
			} else {
				redir := uni.Req.Referer()
				if red, ok := uni.Dat["redirect"]; ok {
					redir = red.(string)
				} else if post_red, okr := uni.Req.Form["redirect"]; okr && len(post_red) == 1 {
					redir = post_red[1]
				}
				http.Redirect(uni.W, uni.Req, redir, 303)
			}
		}
	
	// A back hookokba iratkozik minden, ami legfelsőbb szintre akar feliratkozni, nem nézet/view.
	func runBackHooks(uni *context.Uni) {
		top_hooks, ok := jsonp.GetS(uni.Opt, "Hooks.Back")
		if ok && len(top_hooks) > 0 {
			for _, v := range top_hooks {
				vs := v.(string)
				if h := mod.GetHook(vs, "Back"); h != nil {
					h(uni)
				} else {
					Put(vs + unexported_back)
					return
				}
				if _, ok := uni.Dat["_hijacked"]; ok {
					handleBacks(uni)
					return
				}
			}
			Put(no_back_hijacked)
		} else {
			Put(back_hook_not_set)
		}
	}
	
	// alap konvenció a dispatchhez valószínűleg az lesz h, "admin", "admin/modulnév"
	func runAdminHooks(uni *context.Uni) {
		if len(uni.Paths) > 2 && uni.Paths[2] == "b" {
			admin.AB(uni)
			handleBacks(uni)
		} else {
			admin.AD(uni)
			display.D(uni)
		}
	}
	
	// használat: /debug/modulnév és lefuttatja a modul tesztjét, ami összehasonlítja a jelenlegi opció állományt az elvárt, "papírforma szerintivel"
	func runDebug(uni *context.Uni) {
		mod.Test(uni, uni.Paths[2])
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
	fmt.Println(uni.Dat["_user"])
	buildUser(uni)
	switch uni.Paths[1] {
		// a backhookot azért hoztuk a "/b" mögé, hogy ne foglalja fölöslegesen a sok modulnév a névteret.
		case "b":
			runBackHooks(uni)
		// admin azért csókos, mert beszart opciókkal is működik így
		case "admin":
			runAdminHooks(uni)
		// debug szintén, szétbaszcsizott opciókkal is megy
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

// Egyszerűsítő függvény, hogy teszteljük hogy egy map tartalmaz-e egy adott kulcsot, plusz ezt még egy mutex lockkal le is védi, hogy párhuzamosan futó
// goroutine-okból is meghívhassuk.
func has(c map[string]string, str string) (interface{}, bool) {
	mut := new(sync.Mutex)
	mut.Lock()
	v, ok := c[str]
	mut.Unlock()
	return v, ok
}

	// Ha bárhol befagy a rendszer, és felbuborékol egészen a getSite-ig, ez kiírja a http válaszban a hibát, és a stack tracet.
	func err() {
        if r := recover(); r != nil && r != "controlled" {
			fmt.Println(r)
           	Put(unfortunate_error)
			Put(fmt.Sprint("\n", r, "\n\n" + string(debug.Stack())))
        } else if r != nil && r == "controlled" {
			fmt.Println(r)
			Put(unfortunate_error)
		}
    }

var cache = make(map[string]string)
// A getSite előszedi az oldalhoz tartozó legfrissebb opciót az options kollecióból, aztán átadja a vezérlést a runSite-nak.
// Cachel is, és ő hozza létre a *context.Uni egy példányát is, amit az egész rendszer használ.
func getSite(db *mgo.Database, w http.ResponseWriter, req *http.Request) {
	Put = func (a ...interface{}) {
		io.WriteString(w, fmt.Sprint(a...) + "\n")
	}
	defer err() 
	host := req.Host
	uni := &context.Uni{
		Db: db,
		W: w,
		Req: req,
		Put: Put,
		Dat: make(map[string]interface{}),
		Root: ABSOLUTE_PATH, P: req.URL.Path,
		Paths: strings.Split(req.URL.Path, "/"),
	}
	if val, ok := has(cache, host); ok {
		var v interface{}
		json.Unmarshal([]byte(val.(string)), &v)
		if v == nil {
			Put(inv_userspace)
			return
		}
		uni.Opt = v.(map[string]interface{})
	} else {
		var res interface{}
		db.C("options").Find(m{"Host": host}).Sort(m{"created":-1}).Limit(1).One(&res)
		if res == nil {
			Put(site_not_found)
			return
		} else {
			s, hasU := res.(bson.M)["Userspace"]
			if hasU {
				str := s.(string)
				set(cache, host, str)
				var v interface{}
				json.Unmarshal([]byte(str), &v)
				if v == nil {
					Put(inv_userspace)
					return
				}
				uni.Opt = v.(map[string]interface{})
			} else {
				Put(userspace_not_set)
				return
			}
		}
	}
	req.ParseForm()
	runSite(uni)
}

func main() {
	flag.Parse()
	fmt.Println("server has started")
	defer func(){
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
	http.ListenAndServe("127.0.0.1:" + PORT_NUM, nil)
}
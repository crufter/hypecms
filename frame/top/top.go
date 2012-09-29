package top

import(
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/hypecms/frame/config"
	//"github.com/opesun/hypecms/frame/shell"
	"github.com/opesun/hypecms/frame/mod"
	"github.com/opesun/hypecms/frame/misc/scut"
	"github.com/opesun/hypecms/frame/display"
	"github.com/opesun/hypecms/frame/lang"
	"github.com/opesun/hypecms/frame/lang/speaker"
	"github.com/opesun/hypecms/frame/filter"
	//iface "github.com/opesun/hypecms/frame/interfaces"
	"net/http"
	"fmt"
	"io"
	"path/filepath"
	"labix.org/v2/mgo"
	"encoding/json"
	"strings"
	"runtime/debug"
)

type m map[string]interface{}

var Put func(...interface{})

const (
	unfortunate_error         = "top: An unfortunate error has happened. We are deeply sorry for the inconvenience."
	unexported_front          = "top: Module %v does not export Front view."
	unexported_action         = "top: Module %v does not export action %v."
	no_user_module_build_hook = "top: User module does not export BuildUser hook."
	no_module_at_action       = "top: Tried to execute action, but no module was specified."
	no_action                 = "top: No action specified when accessing module %v."
	no_admin_action		      = "top: No admin action specified."
)

// All front hooks must have the signature of func(*context.Uni, *bool) error
// All views are going to use this hook.
func (t *Top) execFrontHooks() {
	var err error
	i := func(hijacked bool, er error) bool {
		if er != nil {
			err = er
			return true
		}
		return hijacked
	}
	t.uni.Ev.Iterate("Front", i)
	if err == nil {
		display.D(t.uni)
	} else {
		display.DErr(t.uni, err)
	}
}

// After running a background operation this either redirects with data in url paramters or prints out the json encoded result.
func (t *Top) actionResponse(err error, action_name string) {
	uni := t.uni
	if t.config.Debug {
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

func sanitize(a string) string {
	a = strings.Replace(a, "-", " ", -1)
	a = strings.Replace(a, "_", " ", -1)
	a = strings.Title(a)
	return strings.Replace(a, " ", "", -1)
}

func (t *Top) buildUser() error {
	var err error
	ret_rec := func(e error) {
		e = err
	}
	ins := t.uni.NewModule("users").Instance()
	ins.Method("Init").Call(nil, t.uni)
	ins.Method("BuildUser").Call(ret_rec)
	return err
}

// Just printing the stack trace to http response if a panic bubbles up all the way to top.
func topErr() {
	if r := recover(); r != nil {
		fmt.Println("main:", r)
		fmt.Println(string(debug.Stack()))
		Put(unfortunate_error)
		Put(fmt.Sprint("\n", r, "\n\n"+string(debug.Stack())))
	}
}

type Top struct{
	uni 	*context.Uni
	config 	*config.Config
}

func hasVerb(a string, b string) bool {
	mo := mod.NewModule(a)
	if !mo.Exists() {
		return false
	}
	return mo.Instance().HasMethod(b)
}

func (t *Top) Route() {
	uni := t.uni
	if t.config.ServeFiles && strings.Index(uni.Paths[len(uni.Paths)-1], ".") != -1 {
		t.serveFile()
		return
	}
	t.uni.Req.ParseForm()		// Should we handle the error return of this?
	err := t.buildUser()
	if err != nil {
		display.DErr(uni, err)
		return
	}
	r, err := lang.InterpretRoute(uni.P, t.uni.Req.Form)
	if err != nil {
		Put(err.Error())
		return
	}
	filters := []*filter.Filter{}
	var data map[string]interface{}
	view := uni.Req.Method == "GET"
	for i, v := range r.Words {
		if len(r.Words) == i+1 && !view {	// Last one.
			data = filter.ToData(r.Queries[i])
		} else {
			filters = append(filters, filter.New(uni.Db, v, filter.ToQuery(r.Queries[i])))
		}
	}
	nouns := map[string]interface{}{}
	if val, has := uni.Opt["nouns"]; has {
		nouns = val.(map[string]interface{})
	}
	fmt.Println("opt:", uni.Opt)
	speak := speaker.New(hasVerb, nouns)
	s := lang.Translate(r, speak)
	loc := speak.VerbLocation(s.Noun, s.Verb)
	f, err := filter.Reduce(filters...)
	if err != nil {
		panic(err)
	}
	if view {
		uni.Dat["_points"] = []string{loc+"/"+s.Verb}
	}
	ins := mod.NewModule(loc).Instance()
	ins.Method("Init").Call(nil, t.uni)
	if view {
		var res []interface{}
		ret_rec := func(result []interface{}, e error) {
			fmt.Println("rezz", result)
			res = result
			e = err
		}
		ins.Method(s.Verb).Call(ret_rec, f)
		uni.Dat[s.Noun] = res
		if err != nil {
			display.DErr(uni, err)
		}
		display.D(uni)
	} else {
		ret_rec := func(e error) {
			e = err
		}
		ins.Method(s.Verb).Call(ret_rec, f, data)
		t.actionResponse(err, s.Verb)
	}
}

func New(session *mgo.Session, db *mgo.Database, w http.ResponseWriter, req *http.Request, config *config.Config) *Top {
	Put = func(a ...interface{}) {
		io.WriteString(w, fmt.Sprint(a...)+"\n")
	}
	defer topErr()
	uni := &context.Uni{
		Db:      db,
		W:       w,
		Req:     req,
		Put:     Put,
		Dat:     make(map[string]interface{}),
		Root:    config.AbsPath,
		P:       req.URL.Path,
		Paths:   strings.Split(req.URL.Path, "/"),
		NewModule:	mod.NewModule,
	}
	// Not sure if not giving the db session to nonadmin installations increases security, but hey, one can never be too cautious, they dont need it anyway.
	if config.DBAdmMode {
		uni.Session = session
	}
	uni.Ev = context.NewEv(uni)
	opt, opt_str, err := queryConfig(uni.Db, req.Host, config.CacheOpt) // Tricky part about the host, see comments at main_model.
	if err != nil {
		Put(err.Error())
		return &Top{}
	}
	uni.Req.Host = scut.Host(req.Host, opt)
	uni.Opt = opt
	uni.SetOriginalOpt(opt_str)
	uni.SetSecret(config.Secret)
	return &Top{uni,config}
}

// Since we don't include the template name into the url, only "template", we have to extract the template name from the opt here.
// Example: xyz.com/template/style.css
//			xyz.com/tpl/admin/style.css
func (t *Top) serveFile() {
	uni := t.uni
	first_p := uni.Paths[1]
	last_p := uni.Paths[len(uni.Paths)-1]
	has_sfx := strings.HasSuffix(last_p, ".go")
	if first_p == "template" || first_p == "tpl" && !has_sfx {
		t.serveTemplateFile()
	} else if !has_sfx {
		if uni.Paths[1] == "shared" {
			http.ServeFile(uni.W, uni.Req, filepath.Join(t.config.AbsPath, uni.Req.URL.Path))
		} else {
			http.ServeFile(uni.W, uni.Req, filepath.Join(t.config.AbsPath, "uploads", uni.Req.Host, uni.Req.URL.Path))
		}
	} else {
		uni.Put("Don't do that.")
	}
}

func (t *Top) serveTemplateFile() {
	uni := t.uni
	if uni.Paths[1] == "template" {
		p := scut.GetTPath(uni.Opt, uni.Req.Host)
		http.ServeFile(uni.W, uni.Req, filepath.Join(uni.Root, p, strings.Join(uni.Paths[2:], "/")))
	} else { // "tpl"
		http.ServeFile(uni.W, uni.Req, filepath.Join(uni.Root, "modules", uni.Paths[2], "tpl", strings.Join(uni.Paths[3:], "/")))
	}
}
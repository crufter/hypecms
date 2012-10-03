package top

import(
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/hypecms/frame/config"
	"github.com/opesun/hypecms/frame/mod"
	"github.com/opesun/hypecms/frame/misc/scut"
	"github.com/opesun/hypecms/frame/display"
	"github.com/opesun/hypecms/frame/lang"
	"github.com/opesun/hypecms/frame/lang/speaker"
	"github.com/opesun/hypecms/frame/filter"
	"github.com/opesun/jsonp"
	"net/http"
	"net/url"
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"strings"
	"runtime/debug"
)

type m map[string]interface{}

var Put func(...interface{})

const (
	unfortunate_error         = "top: An unfortunate error has happened. We are deeply sorry for the inconvenience."
	no_user_module_build_hook = "top: User module does not export BuildUser hook."
)

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
	nouns := map[string]interface{}{}
	if val, has := uni.Opt["nouns"]; has {
		nouns = val.(map[string]interface{})
	}
	speak := speaker.New(hasVerb, nouns)
	s := lang.Translate(r, speak)
	_, has := jsonp.Get(uni.Opt, "nouns."+s.Noun)
	if !has {
		Put(fmt.Sprintf("Noun %v is undefined.", s.Noun))
		return
	}
	for i, v := range r.Words {
		if len(r.Words) == i+1 && s.Verb != "Get" {	// Last one.
			data = filter.ToData(r.Queries[i])
		} else {
			filters = append(filters, filter.New(uni.Db, v, filter.ToQuery(r.Queries[i])))
		}
	}
	uni.R = r
	uni.S = s
	loc := speak.VerbLocation(s.Noun, s.Verb)
	f, err := filter.Reduce(filters...)
	if err != nil {
		panic(err)
	}
	view := uni.Req.Method == "GET"
	if view {
		uni.Dat["_points"] = []string{loc+"/"+s.Verb}
	}
	uni.Dat["main_noun"] = s.Noun
	ins := mod.NewModule(loc).Instance()
	ins.Method("Init").Call(nil, t.uni)
	if view {
		var ret_rec interface{}
		if s.Verb == "Get" {
			ret_rec = func(res []interface{}, e error) {
				uni.Dat["main"] = res
				e = err
			}
		} else {
			ret_rec = func(e error) {
				e = err
			}
		}
		ins.Method(s.Verb).Call(ret_rec, f)
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

func Modifiers(a url.Values) map[string]interface{} {
	flags := []string{"json", "src", "nofmt", "ok", "action"}
	mods := map[string]interface{}{}
	for _, v := range flags {
		if val, has := a[v]; has {
			mods[v] = val
			delete(a, v)
		}
	}
	for i, v := range a {
		if i[0] == '-' {
			mods[i[1:]] = v
			delete(a, i)
		}
	}
	return mods
}

func New(session *mgo.Session, db *mgo.Database, w http.ResponseWriter, req *http.Request, config *config.Config) *Top {
	Put = func(a ...interface{}) {
		io.WriteString(w, fmt.Sprint(a...)+"\n")
	}
	defer topErr()
	uni := &context.Uni{
		Db:      	db,
		W:       	w,
		Req:     	req,
		Put:     	Put,
		Dat:     	make(map[string]interface{}),
		Root:    	config.AbsPath,
		P:       	req.URL.Path,
		Paths:   	strings.Split(req.URL.Path, "/"),
		NewModule:	mod.NewModule,
	}
	uni.Req.ParseForm()		// Should we handle the error return of this?
	mods := Modifiers(uni.Req.Form)
	uni.Modifiers = mods
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
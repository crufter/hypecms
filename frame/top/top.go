package top

import(
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/hypecms/frame/config"
	"github.com/opesun/hypecms/frame/shell"
	"github.com/opesun/hypecms/frame/mod"
	"github.com/opesun/hypecms/frame/misc/scut"
	"github.com/opesun/hypecms/frame/display"
	//"github.com/opesun/hypecms/frame/filter"
	iface "github.com/opesun/hypecms/frame/interfaces"
	"github.com/opesun/hypecms/modules/users"
	"net/http"
	"fmt"
	"io"
	"path/filepath"
	"labix.org/v2/mgo"
	"encoding/json"
	"strings"
	"net/url"
	"runtime/debug"
	"strconv"
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

func (t *Top) runAction() (string, error) {
	uni := t.uni
	l := len(uni.Paths)
	if l < 3 {
		return "", fmt.Errorf(no_module_at_action)
	}
	modname := uni.Paths[2] // TODO: Routing based on Paths won't work if the site is installed to subfolder or something.
	if l < 4 {
		return "", fmt.Errorf(no_action, modname)
	}
	action_name := uni.Paths[3]
	err, puzzle_err := users.OkayToDoAction(uni, modname, action_name)
	if err != nil {
		return action_name, err
	}
	if puzzle_err != nil {
		return action_name, puzzle_err
	}
	sanitized_aname := sanitize(action_name)
	if !uni.Caller.Has(modname, sanitized_aname) {
		return action_name, fmt.Errorf(unexported_action, modname)
	}
	if !uni.Caller.Matches(modname, sanitized_aname, func() error {return nil}) {
		return action_name, fmt.Errorf("Action %v of %v has bad signature.", action_name, modname)
	}
	ret_rec := func(e error){
		err = e
	}
	uni.Caller.Call(modname, sanitized_aname, ret_rec)
	return action_name, err
}

func (t *Top) execAction() {
	action_name, err := t.runAction()
	t.actionResponse(err, action_name)
}

func (t *Top) buildUser() error {
	// Why is this a hook? Get rid of it.
	if !t.uni.Caller.Has("users", "BuildUser") {
		return fmt.Errorf(no_user_module_build_hook)
	}
	var err error
	ret_rec := func(e error) {
		e = err
	}
	t.uni.Caller.Call("users", "BuildUser", ret_rec)
	return err
}

func (t *Top) terminal() {
	shell.Terminal(t.uni)
	display.D(t.uni)
}

func (t *Top) execCommands() {
	err := shell.FromWeb(t.uni)
	t.actionResponse(err, "shell")
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


type Route struct {
	checked			int
	Words			[]string
	Queries			[]url.Values
}

func (r *Route) Get() string {
	r.checked++
	return r.Words[len(r.Words)-r.checked]
}

func (r *Route) Got() int {
	return r.checked
}

func (r *Route) DropOne() {
	r.Words = r.Words[:len(r.Words)-1]
	r.Queries = r.Queries[:len(r.Queries)-1]
}

func (r *Route) HasMorePair() bool {
	return len(r.Words)>=2+r.checked
}

func sortParams(q url.Values) map[int]url.Values {
	sorted := map[int]url.Values{}
	for i, v := range q {
		num, err := strconv.Atoi(string(i[0]))
		nummed := false
		if err == nil {
			nummed = true
		} else {
			num = 0
		}
		if nummed {
			i = i[1:]
		}
		if _, has := sorted[num]; !has {
			sorted[num] = url.Values{}
		}
		for _, x := range v {
			sorted[num].Add(i, x)
		}
	}
	return sorted
}

// New		Post
// Edit		Put							
func InterpretRoute(p string, q url.Values) (*Route, error) {
	ps := strings.Split(p, "/")
	r := &Route{}
	r.Queries = []url.Values{}
	r.Words = []string{}
	if len(ps) < 1 {
		return r, fmt.Errorf("Wtf.")
	}
	ps = ps[1:]		// First one is empty string.
	sorted := sortParams(q)
	skipped := 0
	for i:=0;i<len(ps);i++ {
		v := ps[i]
		r.Words = append(r.Words, v)
		r.Queries = append(r.Queries, url.Values{})
		qi := len(r.Words)-1
		if len(ps) > i+1 {	// We are not at the end.
			next := ps[i+1]
			if next[1] == '-' && next[0] == v[0] {	// Id query in url., eg /users/u-fxARrttgFd34xdv7
				skipped++
				r.Queries[qi].Add("id", strings.Split(next, "-")[1])
				i++
				continue
			}
		}
		r.Queries[qi] = sorted[qi-skipped]
	}
	return r, nil
}

type Sentence struct{
	Noun, Verb, Redundant string
}

func Translate(r *Route, a iface.Caller, opt map[string]interface{}) *Sentence {
	s := &Sentence{}
	if len(r.Words) == 1 {
		s.Noun = r.Words[0]
		s.Verb = "Get"
		return s
	}
	instable := r.Get()
	must_be_noun := r.Get()
	if a.Exists(instable) {
		s.Verb = "Get"
	} else if a.Has(must_be_noun, instable) {
		s.Verb = instable
	} else {
		s.Redundant = instable
		r.DropOne()
		s.Verb = "Get"
	}
	if !a.Exists(must_be_noun) {
		panic("Noun %v does not exist.")
	}
	s.Noun = must_be_noun
	return s
}

func (t *Top) Route() {
	uni := t.uni
	if t.config.ServeFiles && strings.Index(uni.Paths[len(uni.Paths)-1], ".") != -1 {
		t.serveFile()
	}
	t.uni.Req.ParseForm()		// Should we handle the error return of this?
	err := t.buildUser()
	if err != nil {
		display.DErr(uni, err)
		return
	}
	r, err := InterpretRoute(uni.P, t.uni.Req.Form)
	if err != nil {
		Put(err.Error())
		return
	}
	sen := Translate(r, uni.Caller, uni.Opt)
	fmt.Println("--r", r)
	fmt.Println("-----sent", sen)
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
	}
	uni.Caller = mod.NewCall(uni)
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
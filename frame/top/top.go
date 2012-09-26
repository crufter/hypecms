package top

import(
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/hypecms/frame/config"
	"github.com/opesun/hypecms/frame/shell"
	"github.com/opesun/hypecms/frame/mod"
	"github.com/opesun/hypecms/frame/misc/scut"
	"github.com/opesun/hypecms/frame/display"
	"github.com/opesun/hypecms/modules/user"
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
	err, puzzle_err := user.OkayToDoAction(uni, modname, action_name)
	if err != nil {
		return action_name, err
	}
	if puzzle_err != nil {
		return action_name, puzzle_err
	}
	sanitized_aname := sanitize(action_name)
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

func (t *Top) execAction() {
	action_name, err := t.runAction()
	t.actionResponse(err, action_name)
}

func (t *Top) adminAction() {
	uni := t.uni
	l := len(uni.Paths)
	var err error
	var action_name string
	if l > 3 {
		action_name = uni.Paths[3]
		if scut.IsAdmin(uni.Dat["_user"]) || action_name == "login" || action_name == "regfirstadmin" {
			ret_rec := func(e error) {
				err = e
			}
			err = uni.Caller.Call("actions", "admin", sanitize(action_name), ret_rec)
		}
	} else {
		err = fmt.Errorf(no_admin_action)
	}
	t.actionResponse(err, action_name)
}

func (t *Top) adminView() {
	uni := t.uni
	l := len(uni.Paths)
	var err error
	if !scut.IsAdmin(uni.Dat["_user"]) {
		err = fmt.Errorf("Not allowed.")
	}
	var view string
	if l > 2 {
		view = uni.Paths[2]
	} else {
		view = "index"
	}
	ret_rec := func(e error) {
		e = err
	}
	sane_view := sanitize(view)
	if uni.Caller.Has("views", "admin", sane_view) {
		if view == "index" {
			uni.Dat["_points"] = []string{"admin/index"}
		}
		uni.Caller.Call("views", "admin", sane_view, ret_rec)
	} else {
		modname := view
		if l > 3 {
			view = uni.Paths[3]
		} else {
			view = "index"
		}
		sane_view = sanitize(view)
		uni.Caller.Call("views", modname, "AdminInit", nil)
		if uni.Caller.Has("views", modname, "Admin"+sane_view) {
			uni.Dat["_points"] = []string{modname+"/admin-"+view}
			uni.Caller.Call("views", modname, "Admin"+sane_view, ret_rec)
		} else {
			uni.Dat["_points"] = []string{modname+"/"+view}
			uni.Caller.Call("views", modname, sane_view, ret_rec)
		}
	}
	if err == nil {
		display.D(uni)
	} else {
		display.DErr(uni, err)
	}
}

func (t *Top) routeAdmin() {
	uni := t.uni
	l := len(uni.Paths)
	if l > 2 && uni.Paths[2] == "b" {
		t.adminAction()
	} else {
		t.adminView()
	}
}

func (t *Top) buildUser() error {
	// Why is this a hook? Get rid of it.
	if !t.uni.Caller.Has("hooks", "user", "BuildUser") {
		return fmt.Errorf(no_user_module_build_hook)
	}
	var err error
	ret_rec := func(e error) {
		e = err
	}
	t.uni.Caller.Call("hooks", "user", "BuildUser", ret_rec)
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

func (t *Top) Route() {
	uni := t.uni
	first_p := uni.Paths[1]
	last_p := uni.Paths[len(uni.Paths)-1]
	if t.config.ServeFiles && strings.Index(last_p, ".") != -1 {
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
		return
	}
	t.uni.Req.ParseForm()		// Should we handle the error return of this?
	err := t.buildUser()
	if err != nil {
		display.DErr(uni, err)
		return
	}
	switch uni.Paths[1] {
	// Back hooks are put behind "/b/" to avoid eating up the namespace.
	case "b":
		t.execAction()
	// Admin is a VIP module, to allow bootstrapping a site even if the option document is empty.
	case "admin":
		t.routeAdmin()
	case "run-commands":
		t.execCommands()
	case "terminal":
		t.terminal()
	default:
		t.execFrontHooks()
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
func (t *Top) serveTemplateFile() {
	uni := t.uni
	if uni.Paths[1] == "template" {
		p := scut.GetTPath(uni.Opt, uni.Req.Host)
		http.ServeFile(uni.W, uni.Req, filepath.Join(uni.Root, p, strings.Join(uni.Paths[2:], "/")))
	} else { // "tpl"
		http.ServeFile(uni.W, uni.Req, filepath.Join(uni.Root, "modules", uni.Paths[2], "tpl", strings.Join(uni.Paths[3:], "/")))
	}
}
package shell

import(
	"fmt"
	"github.com/opesun/extract"
	"github.com/opesun/jsonp"
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/hypecms/frame/interfaces"
	"github.com/opesun/hypecms/frame/misc/scut"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"strings"
	"bytes"
	"strconv"
	"text/template"
	"reflect"
)

const done = "Done."

func toString(i interface{}) string {
	switch t := i.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', 5, 64)
	case bool:
		if t == true {
			return "true"
		} else {
			return "false"
		}
	}
	return ""
}

// This is very stupid atm.
func toQueryString(m map[string]interface{}, values url.Values) {
	for i, v := range m {
		switch t := v.(type) {
		case []interface{}:
			for _, x := range t {
				values.Add(i, toString(x))
			}
		default:
			values.Add(i, toString(t))
		}
	}
}

func doParams(params interface{}) (url.Values, error) {
	var m map[string]interface{}
	values := url.Values{}
	switch t := params.(type) {
	case string:
		var i interface{}
		fmt.Println(t)
		err := json.Unmarshal([]byte(t), &i)
		if err != nil {
			return values, err
		}
		m = i.(map[string]interface{})
	case map[string]interface{}:
		m = t
	default:
		return values, fmt.Errorf("doParams: Unkown type.")
	}
	toQueryString(m, values)
	return values, nil
}

type ma map[string]interface{}

func do(uni *context.Uni, module, action string, params interface{}) map[string]interface{} {
	values, err := doParams(params)
	if err != nil {
		return ma{"error": err}
	}
	var url_path string
	if module == "admin" {
		url_path = fmt.Sprintf("/admin/b/%v", action)
	} else {
		_, has := jsonp.GetM(uni.Opt, "Modules." + module)
		if !has {
			return ma{"error": fmt.Sprintf("Module %v is not installed.", module)}
		}
		url_path = fmt.Sprintf("/b/%v/%v", module, action)
	}
	values.Add("json", "true")
	full_body := values.Encode()
	req, err := http.NewRequest("POST", "http://"+uni.Req.Host+url_path, bytes.NewBufferString(full_body))
	if err != nil {
		return ma{"error": err}
	}
	if cookie, err := uni.Req.Cookie("user"); err == nil {
		req.AddCookie(cookie)
	}
	req.Header.Add("Content-Length", strconv.Itoa(len(full_body)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := new(http.Client).Do(req)
	if err != nil {
		return ma{"error": err}
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ma{"error": err}
	}
	var i interface{}
	err = json.Unmarshal(contents, &i)
	if err != nil {
		return ma{"error": "do request response is not a valid json."}		// Happens when the requested page panics.
	}
	ret := i.(map[string]interface{})
	err_val, has := ret["error"]
	if has {
		return ma{"error": err_val.(string)}
	}
	return ret
}

func keys(m map[string]interface{}) []interface{} {
	ret := []interface{}{}
	for i := range m {
		ret = append(ret, i)
	}
	return ret
}

func concat(a ...string) string {
	return strings.Join(a, "")
}

func musth(a string, b map[string]interface{}) error {
	_, has := jsonp.Get(b, a)
	if !has {
		return fmt.Errorf("Map has no key \"%v\". Terminating.", a)
	}
	return nil
}

var arg_labels = []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z"}

// Help creates a string describing the signature of a given function for documentation purposes.
// There must be a better way for sure.
func help(funcmap, docmap map[string]interface{}, func_name string) string {
	fun, has := funcmap[func_name]
	if !has {
		return fmt.Sprintf("No function named %v.", func_name)
	}
	v := reflect.TypeOf(fun)
	ret := ""
	ret = ret + fmt.Sprintf("\nfunc %v(", func_name)
	for i:=0;i<v.NumIn();i++{
		ret = ret+arg_labels[i]+" "+fmt.Sprint(v.In(i))
		if i<v.NumIn()-1 {
			ret = ret + ", "
		}
	}
	ret = ret + ") "
	if v.NumOut() > 1 {
		ret = ret + "("
	}
	for i:=0;i<v.NumOut();i++{
		ret = ret + fmt.Sprint(v.Out(i))
		if i<v.NumOut()-1 {
			ret = ret + ", "
		}
	}
	if v.NumOut() > 1 {
		ret = ret + ")"
	}
	ret = ret+"\n"+docmap[func_name].(string)
	return ret
}

func commands(b map[string]interface{}) []string {
	ret := []string{}
	for i := range b {
		ret = append(ret, i)
	}
	return ret
}

func avail(a map[string]interface{}, b string) bool {
	_, has := a[b]
	return has
}

func sil(a interface{}) error {
	return nil
}

func names(a interfaces.Caller, modname string) []string {
	return a.Names(modname)
}

// The way the documentation is provided may become subject to change, because it's ugly as hell.
func builtins(uni *context.Uni) map[string]interface{} {
	d := map[string]interface{}{
		"do": 			"Do calls action 'b' of module 'a' with the given POST params 'c'.\nJSON is used for params because it is easier to write and read than raw query strings.\n\nDo is your swiss army knife.",
		"install":		"Installs a module.",
		"uninstall":	"Uninstalls a module.",
		"keys":			"Returns a list of keys in map 'a'.",
		"concat":		"Concatenates an arbitrary number of strings.",
		"musth":		`Abr. of "must have". Panics if the map 'b' has no field named 'a'.`,
		"sil":			`Silences the output of another function.\nExample usage:\ninstall "content" | sil`,
		"actions":		"Returns the name of all actions of a module. Module may be uninstalled.",
		"views":		"Returns the name of all views of a module. Module may be uninstalled.",
		"hooks":		"Returns the name of all hooks of a module. Module may be uninstalled.",
	}
	f := map[string]interface{}{ 
		"do": func(a, b, c string) map[string]interface{} {
			return do(uni, a, b, c)
		},
		"install": func(a string) map[string]interface{} {
			return do(uni, "admin", "install", `{"module":"`+a+`"}`)
		},
		"uninstall": func(a string) map[string]interface{} {
			return do(uni, "admin", "uninstall", `{"module":"`+a+`"}`)
		},
		"keys": keys,
		"concat": concat,
		"musth": musth,
		"sil": sil,
		"exported": func(a string) []string {
			return names(uni.Caller, a)
		},
	}
	d["commands"] = "Returns a list of all command names."
	d["help"] = "Displays the help for a given command."
	d["avail"] = "Returns true if the given command is available."
	f["commands"] = func() []string { return commands(f) }
	f["help"] = func(a string) string { return help(f, d, a) }
	f["avail"] = func(a string) bool { return avail(f, a) }
	uni.Ev.Trigger("shellFunctions", uni, f, d)
	return f
}

func stripComments(lines []string) []string {
	ret := []string{}
	for _, v := range lines {
		if len(v) > 0 && v[0] != '#' {
			ret = append(ret, v)
		}
	}
	return ret
}

func strip(e error) error {
	return fmt.Errorf("line "+e.Error()[16:])
}

func Run(uni *context.Uni, commands string) (string, error) {
	if !scut.IsAdmin(uni.Dat["_user"]) {
		return "", fmt.Errorf("Currently only admins can use this.")
	}
	lines := strings.Split(commands, "\n")
	for i, v := range lines {
		// This hack...
		if len(v) > 0 && v[0] != '#' {
			lines[i] = "{{"+v+"}}"
		}
	}
	whole_file := strings.Join(lines, "\n")
	funcMap := template.FuncMap(builtins(uni))
	t, err := template.New("shell").Funcs(funcMap).Parse(string(whole_file))
	if err != nil {
		return "", strip(err)
	}
	context := map[string]interface{}{"dat": uni.Dat, "opt": uni.Opt}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, context) // TODO: watch for errors in execution.
	if err != nil {
		return "", strip(err)
	}
	output_lines := strings.Split(buffer.String(), "\n")
	output_lines = stripComments(output_lines)
	return strings.Join(output_lines, "\n"), nil
}

func Terminal(uni *context.Uni) error {
	if !scut.IsAdmin(uni.Dat["_user"]) {
		return fmt.Errorf("Currently only admins can use this.")
	}
	uni.Dat["_points"] = []string{"admin/terminal"}		// This is inconsequential with other parts of the system, probably there is a better place for this.
	return nil
}

func FromWeb(uni *context.Uni) error {
	dat, err := extract.New(map[string]interface{}{"commands": "must"}).Extract(uni.Req.Form)
	if err != nil {
		return err
	}
	res, err := Run(uni, dat["commands"].(string))
	if err != nil {
		return err
	}
	uni.Dat["_cont"] = map[string]interface{}{"output": res}
	return nil
}
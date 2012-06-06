// Module Display will execute the given "Display Points" ([]string found in uni.Dat["_points"]).
// If there is no "_points" set in uni.Dat, then it will execute the Display Point which name matches the http Request Path.
// A given Display Point can contain queries, they will be run, after that, a tpl file which matches the name of the Display Point will be executed.
//
// TODO: use error troughout the package instead of a string error.
// Caution: there is some real tricky (ugly?) hackage going on in this package in at getFile/getTPath/getModPath to support the logic behind the fallback files.
package display

import (
	"encoding/json"
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/require"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func displErr(uni *context.Uni) {
	r := recover()
	if r != nil {
		uni.Put("The template file is a pile of shit and contains malformed data what the html/template pkg can't handle.")
	}
}

// TODO: Implement file caching here.
func getFile(abs, fi string, uni *context.Uni) ([]byte, error) {
	p := getTPath(fi, uni)
	fmt.Println(p)
	b, err := ioutil.ReadFile(filepath.Join(p[0], p[1]))
	if err == nil {
		return b, nil
	}
	p = getModPath(fi, uni)
	fmt.Println(p)
	return ioutil.ReadFile(filepath.Join(p[0], p[1]))
}

func templateType(opt map[string]interface{}) string {
	_, priv := opt["TplIsPrivate"]
	var ttype string
	if priv {
		ttype = "private"
	} else {
		ttype = "public"
	}
	return ttype
}

func templateName(opt map[string]interface{}) string {
	tpl, has_tpl := opt["Template"]
	if !has_tpl {
		tpl = "default"
	}
	return tpl.(string)
}

// [0]: abs_folder, [1]: relative_path
func getTPath(s string, uni *context.Uni) []string {
	sl := []string{}
	templ := templateName(uni.Opt)
	ttype := templateType(uni.Opt)
	sl = append(sl, filepath.Join(uni.Root, "templates", ttype, templ))
	sl = append(sl, s)
	return sl
}

// [0]: abs_folder, [1]: relative_path
func getModPath(s string, uni *context.Uni) []string {
	sl := []string{}
	p := strings.Split(s, "/")
	sl = append(sl, filepath.Join(uni.Root, "modules", p[0], "tpl"))
	sl = append(sl, strings.Join(p[1:], "/"))
	return sl
}

// Executes filep.tpl of a given template.
func DisplayTemplate(uni *context.Uni, filep string) string {
	p := getTPath(filep+".tpl", uni)
	file, err := require.R(p[0], p[1],
		func(abs, fi string) ([]byte, error) {
			return getFile(abs, fi, uni)
		})
	if err == "" {
		uni.Dat["_tpl"] = "/templates/" + templateType(uni.Opt) + "/" + templateName(uni.Opt) + "/"
		t, _ := template.New("template_name").Parse(string(file))
		_ = t.Execute(uni.W, uni.Dat)
		return ""
	}
	return "cant find template file " + `"` + filep + `"`
}

// If a given .tpl can not be found in the template folder, it will try identify the module which can have that .tpl file. 
func DisplayFallback(uni *context.Uni, filep string) string {
	if strings.Index(filep, "/") != -1 {
		if len(strings.Split(filep, "/")) >= 2 {
			p := getModPath(filep+".tpl", uni)
			file, err := require.R(p[0], filep,			// Tricky, care.
				func(abs, fi string) ([]byte, error) {
					return getFile(abs, fi, uni)
				})
			if err == "" {
				uni.Dat["_tpl"] = "/modules/" + p[0] + "/tpl/"
				t, _ := template.New("template_name").Parse(string(file))
				_ = t.Execute(uni.W, uni.Dat)
				return ""
			}
			return "cant find fallback template file " + `"` + filep + `"`
		}
		return "fallback filep is too long"
	}
	return "fallback filep contains no slash, so there nothing to fall back"
}

func DisplayFile(uni *context.Uni, filep string) {
	defer displErr(uni)
	err := DisplayTemplate(uni, filep)
	if err != "" {
		err_f := DisplayFallback(uni, filep)
		if err_f != "" {
			uni.Put(err, "\n", err_f)
		}
	}
}

func queryErr(uni *context.Uni) {
	r := recover()
	fmt.Println(r)
	uni.Put("shit happened while running queries")
}

// Runs the queries associated with a given Display Point.
func runQueries(uni *context.Uni, queries []map[string]interface{}) {
	defer queryErr(uni)
	qs := make(map[string]interface{})
	for _, v := range queries {
		q := uni.Db.C(v["c"].(string)).Find(v["q"])
		if skip, skok := v["sk"]; skok {
			q.Skip(skip.(int))
		}
		if limit, lok := v["l"]; lok {
			q.Limit(limit.(int))
		}
		if sort, sook := v["so"]; sook {
			q.Sort(jsonp.ToStringSlice(sort))
		}
		var res interface{}
		q.All(&res)
		qs[v["n"].(string)] = res
	}
	uni.Dat["queries"] = qs
}

// This is where this module starts.
func D(uni *context.Uni) {
	points, points_exist := uni.Dat["_points"]
	var point, filep string
	if points_exist {
		point = points.([]string)[0]
		queries, queries_exists := jsonp.Get(uni.Opt, "Modules.display.Points."+point+".queries")
		if queries_exists {
			qslice, ok := queries.([]map[string]interface{})
			if ok {
				runQueries(uni, qslice)
			}
		}
		filep = point
		// Ha nincs point
	} else {
		p := uni.Req.URL.Path
		if p == "/" {
			filep = "index"
		} else {
			filep = p
		}
	}
	if _, isjson := uni.Req.Form["json"]; isjson {
		var v []byte
		if _, fmt := uni.Req.Form["fmt"]; fmt {
			v, _ = json.MarshalIndent(uni.Dat["_cont"], "", "    ")
		} else {
			v, _ = json.Marshal(uni.Dat["_cont"])
		}
		uni.Put(v)
	} else {
		DisplayFile(uni, filep)
	}
}

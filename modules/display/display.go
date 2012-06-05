// Module Display will execute the given "Display Points" ([]string found in uni.Dat["_points"]).
// If there is no "_points" set in uni.Dat, then it will execute the Display Point which name matches the http Request Path.
// A given Display Point can contain queries, they will be run, after that, a tpl file which matches the name of the Display Point will be executed.
//
// TODO: use error troughout the package instead of a string error.
package display

import (
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/require"
	"html/template"
	"path/filepath"
	"strings"
	//"io/ioutil"
)

func displErr(uni *context.Uni) {
	r := recover()
	if r != nil {
		uni.Put("The template file is a pile of shit and contains malformed data what the html/template pkg can't handle.")
	}
}

// Executes filep.tpl of a given template.
func DisplayTemplate(uni *context.Uni, filep string) string {
	tpl, has_tpl := uni.Opt["Template"]
	if !has_tpl {
		tpl = "default"
	}
	templ := tpl.(string)
	_, priv := uni.Opt["TplIsPrivate"]
	var ttype string
	if priv {
		ttype = "private"
	} else {
		ttype = "public"
	}
	file, err := require.RSimple(filepath.Join(uni.Root, "templates", ttype, templ), filep+".tpl")
	if err == "" {
		uni.Dat["_tpl"] = "/templates/" + ttype + "/" + templ + "/"
		t, _ := template.New("template_name").Parse(string(file))
		_ = t.Execute(uni.W, uni.Dat)
		return ""
	}
	return "cant find template file " + `"` + filep + `"`
}

// If a given .tpl can not be found in the template folder, it will try identify the module which can have that .tpl file. 
func DisplayFallback(uni *context.Uni, filep string) string {
	if strings.Index(filep, "/") != -1 {
		p := strings.Split(filep, "/")
		if len(p) >= 2 {
			file, err := require.RSimple(filepath.Join(uni.Root, "modules", p[0], "tpl"), strings.Join(p[1:], "/")+".tpl")
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
		queries, queries_exists := jsonp.Get(uni.Opt, "Modules.Display.Points."+point+".queries")
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
	DisplayFile(uni, filep)
}
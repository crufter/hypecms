// Module Display will execute the given "Display Points" ([]string found in uni.Dat["_points"]).
// If there is no "_points" set in uni.Dat, then it will execute the Display Point which name matches the http Request Path.
// A given Display Point can contain queries, they will be run, after that, a tpl file which matches the name of the Display Point will be executed.
//
// Caution: there is some real tricky (ugly?) hackage going on in this package at getFile/getTPath/getModTPath to support the logic behind the fallback files.
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
	"runtime/debug"
	"github.com/opesun/hypecms/api/scut"
)

func displErr(uni *context.Uni) {
	r := recover()
	if r != nil {
		uni.Put("There was an error executing the template. Probably a {{require ...}} remained in the template and the html/template module does not recognize unkown commands, it crashes only.")
		fmt.Println("problem with template: ", r)
		debug.PrintStack()
	}
}

// TODO: Implement file caching here.
func getFile(root, fi string, opt map[string]interface{}) ([]byte, error) {
	p := scut.GetTPath(opt)
	b, err := ioutil.ReadFile(filepath.Join(root, p, fi))
	if err == nil {
		return b, nil
	}
	mp := getModTPath(fi)
	return ioutil.ReadFile(filepath.Join(root, mp[0], mp[1]))
}

// Inp:	"admin/this/that.txt"
// []string{ "modules/admin/tpl", "this/that.txt"}
func getModTPath(filename string) []string {
	sl := []string{}
	p := strings.Split(filename, "/")
	sl = append(sl, filepath.Join("modules", p[0], "tpl"))
	sl = append(sl, strings.Join(p[1:], "/"))
	return sl
}

// Executes filep.tpl of a given template.
func DisplayTemplate(uni *context.Uni, filep string) error {
	file, err := require.R("", filep+".tpl",
		func(root, fi string) ([]byte, error) {
			return getFile(uni.Root, fi, uni.Opt)
		})
	if err == nil {
		uni.Dat["_tpl"] = "/templates/" + scut.TemplateType(uni.Opt) + "/" + scut.TemplateName(uni.Opt) + "/"
		t, _ := template.New("template_name").Parse(string(file))
		_ = t.Execute(uni.W, uni.Dat)
		return nil
	}
	return fmt.Errorf("cant find template file ", `"`, filep, `"`)
}

// If a given .tpl can not be found in the template folder, it will try identify the module which can have that .tpl file. 
func DisplayFallback(uni *context.Uni, filep string) error {
	if strings.Index(filep, "/") != -1 {
		if len(strings.Split(filep, "/")) >= 2 {
			file, err := require.R("", filep + ".tpl",			// Tricky, care.
				func(root, fi string) ([]byte, error) {
					return getFile(uni.Root, fi, uni.Opt)
				})
			if err == nil {
				uni.Dat["_tpl"] = "/modules/" + strings.Split(filep, "/")[0] + "/tpl/"
				t, _ := template.New("template_name").Parse(string(file))
				_ = t.Execute(uni.W, uni.Dat)
				return nil
			}
			return fmt.Errorf("cant find fallback template file ", `"`, filep, `"`)
		}
		return fmt.Errorf("fallback filep is too long")
	}
	return fmt.Errorf("fallback filep contains no slash, so there nothing to fall back")
}

func DisplayFile(uni *context.Uni, filep string) error {
	defer displErr(uni)
	err := DisplayTemplate(uni, filep)
	if err != nil {
		err_f := DisplayFallback(uni, filep)
		if err_f != nil {
			return fmt.Errorf("Can't find file or fallback.")
		}
	}
	return nil
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

// This is where the module starts if an error occured in a front hook.
func DErr(uni *context.Uni, err error) {
	uni.Put(err)
}

// This is where the module starts if there were no error in the front hook.
func D(uni *context.Uni) {
	points, points_exist := uni.Dat["_points"]
	var point, filep string		// filep = file path
	if points_exist {
		point = points.([]string)[0]
		queries, queries_exists := jsonp.Get(uni.Opt, "Display-points." + point + ".queries")
		if queries_exists {
			qslice, ok := queries.([]map[string]interface{})
			if ok {
				runQueries(uni, qslice)
			}
		}
		filep = point
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
		if _, format := uni.Req.Form["fmt"]; format {
			v, _ = json.MarshalIndent(uni.Dat, "", "    ")
		} else {
			v, _ = json.Marshal(uni.Dat)
		}
		uni.Put(string(v))
	} else {
		err := DisplayFile(uni, filep)
		if err != nil {
			uni.Dat["missing_file"] = filep
			err_404 := DisplayFile(uni, "404")
			if err_404 != nil {
				uni.Put("Cant find file: ", filep)
			}
		}
	}
}
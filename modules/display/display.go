// Module Display will execute the given "Display Points" ([]string found in uni.Dat["_points"]).
// If there is no "_points" set in uni.Dat, then it will execute the Display Point which name matches the http Request Path.
// A given Display Point can contain queries, they will be run, after that, a tpl file which matches the name of the Display Point will be executed as a Go template.
//
// Caution: there is some slightly tricky (ugly?) hackage going on in this package at getFile/getTPath/getModTPath to support the logic behind the fallback files.
package display

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/require"
	"github.com/opesun/hypecms/modules/display/model"
	"github.com/opesun/hypecms/model/scut"
	"html/template"
	"strings"
	"runtime/debug"
	"encoding/json"
	"fmt"
)

// Prints errors during template file display to http response. (Kinda nonsense now as it is.)
func displErr(uni *context.Uni) {
	r := recover()
	if r != nil {
		uni.Put("There was an error executing the template. Probably a {{require ...}} remained in the template and the html/template module does not recognize unkown commands, it crashes only.")
		fmt.Println("problem with template: ", r)
		debug.PrintStack()
	}
}

func merge(a interface{}, b map[string]interface{}) map[string]interface{} {
	if a == nil { return b }
	if b == nil { return a.(map[string]interface{}) }
	a_m := a.(map[string]interface{})
	for i, v := range b {
		a_m[i] = v
	}
	return a_m
}

func toStringSlice(a interface{}) []string {
	if a == nil { return nil }
	switch val := a.(type) {
	case []interface{}:
		return jsonp.ToStringSlice(val)
	case []string:
		return val
	}
	return nil
}

func get(dat map[string]interface{}, s ...string) interface{} {
	if len(s) > 0 {
		if len(s[0]) > 0 {
			if string(s[0][0]) == "$" {
				s[0] = s[0][1:]
			}
		}
	}
	access := strings.Join(s, ".")
	val, has := jsonp.Get(dat, access)
	if !has { return access }
	return val
}

// Tries to dislay a template file.
func DisplayTemplate(uni *context.Uni, filep string) error {
	file, err := require.R("", filep+".tpl",
		func(root, fi string) ([]byte, error) {
			return scut.GetFile(uni.Root, fi, uni.Opt, uni.Req.Host, nil)
		})
	if err == nil {
		uni.Dat["_tpl"] = "/templates/" + scut.TemplateType(uni.Opt) + "/" + scut.TemplateName(uni.Opt) + "/"
		langs, has := jsonp.Get(uni.Dat, "_user.languages")																		// _user always has language member
		if !has { langs = []string{"en"} }
		langs_s := toStringSlice(langs)
		loc, _ := display_model.LoadLocTempl(string(file), langs_s, uni.Root, scut.GetTPath(uni.Opt, uni.Req.Host), nil)		// TODO: think about errors here.
		uni.Dat["loc"] = merge(uni.Dat["loc"], loc)
		funcMap := template.FuncMap{
			"get": func(s ...string) interface{} {
				return get(uni.Dat, s...)
			},
		}
		t, _ := template.New("template_name").Funcs(funcMap).Parse(string(file))
		t.Execute(uni.W, uni.Dat)	// TODO: watch for errors in execution.
		return nil
	}
	return fmt.Errorf("cant find template file ", `"`, filep, `"`)
}

// Tries to display a module file.
func DisplayFallback(uni *context.Uni, filep string) error {
	if strings.Index(filep, "/") != -1 {
		if scut.PossibleModPath(filep) {
			file, err := require.R("", filep + ".tpl",			// Tricky, care.
				func(root, fi string) ([]byte, error) {
					return scut.GetFile(uni.Root, fi, uni.Opt, uni.Req.Host, nil)
				})
			if err == nil {
				uni.Dat["_tpl"] = "/modules/" + strings.Split(filep, "/")[0] + "/tpl/"
				langs, has := jsonp.Get(uni.Dat, "_user.languages")																				// _user always has language member
				if !has { langs = []string{"en"} }
				langs_s := toStringSlice(langs)
				if !has { langs = []string{"en"} }
				loc, _ := display_model.LoadLocTempl(string(file), langs_s, uni.Root, scut.GetTPath(uni.Opt, uni.Req.Host), nil)			// TODO: think about errors here.
				uni.Dat["loc"] = merge(uni.Dat["loc"], loc)
				funcMap := template.FuncMap{
					"get": func(s ...string) interface{} {
						return get(uni.Dat, s...)
					},
				}
				t, _ := template.New("template_name").Funcs(funcMap).Parse(string(file))
				t.Execute(uni.W, uni.Dat)	// TODO: watch for errors in execution.
				return nil
			}
			return fmt.Errorf("cant find fallback template file ", `"`, filep, `"`)
		}
		return fmt.Errorf("fallback filep is too long")
	}
	return fmt.Errorf("fallback filep contains no slash, so there nothing to fall back")
}

// Tries to display the relative filepath filep as either a template file or a module file.
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

// Prints errors during query running to http response. (Kinda nonsense now as it is.)
func queryErr(uni *context.Uni) {
	r := recover()
	if r != nil {
		fmt.Println(r)
		uni.Put("There was an error running the queries: ", r)
		debug.PrintStack()
	}
}

// Runs the queries associated with a given Display Point.
func runQueries(uni *context.Uni, queries map[string]interface{}) {
	defer queryErr(uni)
	uni.Dat["queries"] = display_model.RunQueries(uni.Db, queries, map[string][]string(uni.Req.Form), uni.P + "?" + uni.Req.URL.RawQuery)
}

// Prints all available data to http response as a JSON.
func putJSON(uni *context.Uni) {
	var v []byte
	if _, format := uni.Req.Form["fmt"]; format {
		v, _ = json.MarshalIndent(uni.Dat, "", "    ")
	} else {
		v, _ = json.Marshal(uni.Dat)
	}
	uni.Put(string(v))
}

// This is called if an error occured in a front hook.
func DErr(uni *context.Uni, err error) {
	if _, isjson := uni.Req.Form["json"]; isjson {
		putJSON(uni)
		return
	}
	uni.Put(err)
}

// Displays a display point.
func D(uni *context.Uni) {
	points, points_exist := uni.Dat["_points"]
	var point string
	if points_exist {
		point = points.([]string)[0]
	} else {
		p := uni.Req.URL.Path
		if p == "/" {
			point = "index"
		} else {
			point = p
		}
	}
	queries, queries_exists := jsonp.Get(uni.Opt, "Display-points." + point + ".queries")
	if queries_exists {
		qmap, ok := queries.(map[string]interface{})
		if ok {
			runQueries(uni, qmap)
		}
	}
	langs, _ := jsonp.Get(uni.Dat, "_user.languages")																	// _user always has language member
	langs_s := toStringSlice(langs)
	loc, _ := display_model.LoadLocStrings(uni.Dat, langs_s, uni.Root, scut.GetTPath(uni.Opt, uni.Req.Host), nil)		// TODO: think about errors here.
	if loc != nil {
		uni.Dat["loc"] = loc
	}
	if _, isjson := uni.Req.Form["json"]; isjson {
		putJSON(uni)
		return
	} else {
		err := DisplayFile(uni, point)
		if err != nil {
			uni.Dat["missing_file"] = point
			err_404 := DisplayFile(uni, "404")
			if err_404 != nil {
				uni.Put("Cant find file: ", point)
			}
		}
	}
}

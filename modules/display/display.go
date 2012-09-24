// Module Display will execute the given "Display Points" ([]string found in uni.Dat["_points"]).
// If there is no "_points" set in uni.Dat, then it will execute the Display Point which name matches the http Request Path.
// A given Display Point can contain queries, they will be run, after that, a tpl file which matches the name of the Display Point will be executed as a Go template.
//
// Caution: there is some slightly tricky (ugly?) hackage going on in this package at getFile/getTPath/getModTPath to support the logic behind the fallback files.
package display

import (
	"encoding/json"
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/display/model"
	"github.com/opesun/jsonp"
	"github.com/opesun/require"
	"github.com/russross/blackfriday"
	"html/template"
	"runtime/debug"
	"strings"
)

// Prints errors during template file display to http response. (Kinda nonsense now as it is.)
func displErr(uni *context.Uni) {
	r := recover()
	if r != nil {
		uni.Put("There was an error executing the template.")
		fmt.Println("problem with template: ", r)
		debug.PrintStack()
	}
}

func merge(a interface{}, b map[string]interface{}) map[string]interface{} {
	if a == nil {
		return b
	}
	if b == nil {
		return a.(map[string]interface{})
	}
	a_m := a.(map[string]interface{})
	for i, v := range b {
		a_m[i] = v
	}
	return a_m
}

func toStringSlice(a interface{}) []string {
	if a == nil {
		return nil
	}
	switch val := a.(type) {
	case []interface{}:
		return jsonp.ToStringSlice(val)
	case []string:
		return val
	}
	return nil
}

func validFormat(format string) bool {
	switch format {
	case "md":
		return true
	}
	return false
}

// Does format conversions.
// Currently only: markdown -> html
func GetFileAndConvert(root, fi string, opt map[string]interface{}, host string, file_reader func(string) ([]byte, error)) ([]byte, error) {
	file, err := scut.GetFile(root, fi, opt, host, file_reader)
	if err != nil {
		return file, err
	}
	spl := strings.Split(fi, ".")
	extension := spl[len(spl)-1]
	get := func(root, fi string) ([]byte, error) {
		return GetFileAndConvert(root, fi, opt, host, file_reader)
	}
	file, err = display_model.Load(opt["Loads"], root, file, get)
	if err != nil {
		return nil, err
	}
	// In tpl files the first line contains the extension information, like "--md". (An entry point can't change it's extension.)
	if extension == "tpl" {
		strfile := string(file)
		newline_pos := strings.Index(strfile, "\n")
		if newline_pos > 3 && validFormat(strfile[2:newline_pos-1]) { // "--" plus at least 1 characer.
			extension = strfile[2 : newline_pos-1]
			file = file[newline_pos:]
		}
	}
	switch extension {
	case "md":
		file = blackfriday.MarkdownCommon(file)
	}
	//file = append([]byte(fmt.Sprintf("<!-- %v/%v. -->", root, fi)), file...)
	//file = append(file, []byte(fmt.Sprintf("<!-- /%v/%v -->", root, fi))...)
	return file, nil
}

// Tries to dislay a template file.
func DisplayTemplate(uni *context.Uni, filep string) error {
	_, src := uni.Req.Form["src"]
	file, err := require.R("", filep+".tpl",
		func(root, fi string) ([]byte, error) {
			return GetFileAndConvert(uni.Root, fi, uni.Opt, uni.Req.Host, nil)
		})
	if err != nil {
		return fmt.Errorf("Cant find template file %v.", filep)
	}
	if src {
		uni.Put(string(file))
		return nil
	}
	uni.Dat["_tpl"] = "/templates/" + scut.TemplateType(uni.Opt) + "/" + scut.TemplateName(uni.Opt) + "/"
	prepareAndExec(uni, string(file))
	return nil
}

// Loads localization, template functions and executes the template.
func prepareAndExec(uni *context.Uni, file string) {
	root := uni.Root
	host := uni.Req.Host
	dat := uni.Dat
	opt := uni.Opt
	w := uni.W
	langs, has := jsonp.Get(dat, "_user.languages") // _user should always has languages field
	if !has {
		langs = []string{"en"}
	}
	langs_s := toStringSlice(langs)
	if !has {
		langs = []string{"en"}
	}
	loc, _ := display_model.LoadLocTempl(file, langs_s, root, scut.GetTPath(opt, host), nil) // TODO: think about errors here.
	dat["loc"] = merge(dat["loc"], loc)
	funcMap := template.FuncMap(builtins(uni))
	t, _ := template.New("tpl").Funcs(funcMap).Parse(string(file))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, dat) // TODO: watch for errors in execution.
}

// Tries to display a module file.
func DisplayFallback(uni *context.Uni, filep string) error {
	_, src := uni.Req.Form["src"]
	if strings.Index(filep, "/") != -1 {
		return fmt.Errorf("Nothing to fall back to.") // No slash in fallback path means no modulename to fall back to.
	}
	if scut.PossibleModPath(filep) {
		return fmt.Errorf("Not a possible fallback path.")
	}
	file, err := require.R("", filep+".tpl", // Tricky, care.
		func(root, fi string) ([]byte, error) {
			return GetFileAndConvert(uni.Root, fi, uni.Opt, uni.Req.Host, nil)
		})
	if err != nil {
		fmt.Errorf("Cant find fallback file %v.", filep)
	}
	if src {
		uni.Put(string(file))
		return nil
	}
	uni.Dat["_tpl"] = "/modules/" + strings.Split(filep, "/")[0] + "/tpl/"
	prepareAndExec(uni, string(file))
	return nil
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
		uni.Put("There was an error running the queries: ", r)
		debug.PrintStack()
	}
}

// Runs the queries associated with a given Display Point.
func runQueries(uni *context.Uni, queries map[string]interface{}) {
	defer queryErr(uni)
	uni.Dat["queries"] = display_model.RunQueries(uni.Db, queries, map[string][]string(uni.Req.Form), uni.P+"?"+uni.Req.URL.RawQuery)
}

// Prints all available data to http response as a JSON.
func putJSON(uni *context.Uni) {
	var v []byte
	if _, format := uni.Req.Form["fmt"]; format {
		v, _ = json.MarshalIndent(uni.Dat, "", "    ")
	} else {
		v, _ = json.Marshal(uni.Dat)
	}
	uni.W.Header().Set("Content-Type", "application/json; charset=utf-8")
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

func BeforeDisplay(uni *context.Uni) {
	defer func(){
		r := recover()
		if r != nil {
			fmt.Println(r)
		}
	}()
	uni.Ev.Trigger("BeforeDisplay")
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
	queries, queries_exists := jsonp.Get(uni.Opt, "Display-points."+point+".queries")
	if queries_exists {
		qmap, ok := queries.(map[string]interface{})
		if ok {
			runQueries(uni, qmap)
		}
	}
	BeforeDisplay(uni)
	// While it is not the cheapest solution to convert bson.ObjectIds to strings here, where we have to iterate trough all values,
	// it is still better than remembering (and forgetting) to convert it at every specific place.
	scut.IdsToStrings(uni.Dat)
	langs, _ := jsonp.Get(uni.Dat, "_user.languages") // _user always has language member
	langs_s := toStringSlice(langs)
	loc, _ := display_model.LoadLocStrings(uni.Dat, langs_s, uni.Root, scut.GetTPath(uni.Opt, uni.Req.Host), nil) // TODO: think about errors here.
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

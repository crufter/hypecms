package display_model

import(
	"labix.org/v2/mgo"
	"github.com/opesun/jsonp"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/paging"
	"github.com/opesun/resolver"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"regexp"
	"fmt"
)

// Cuts a long string at max_char_count, taking a word boundary into account.
func Excerpt(s string, max_char_count int) string {
	if len(s) < max_char_count { return s }
	ind := strings.LastIndex(s[:max_char_count], " ")
	if ind == -1 { return s }
	return s[0:ind]
	
}

// Extracts the key value pair of a map which has a length of 1.
func GetOnlyPair(c map[string]interface{}) (string, interface{}) {
	for i, v := range c {
		return i, v
	}
	return "", nil
}

// conf := map[string]interface{}{"content": 
// Maybe we could modify this to be able to create excerpts from multiple fields.
func CreateExcerpts(res []interface{}, conf map[string]interface{}) {
	fieldname, max_char_i := GetOnlyPair(conf)
	max_char := toInt(max_char_i)
	for _, v := range res {
		doc := v.(map[string]interface{})
		field_val, ok := doc[fieldname].(string)
		if !ok { continue }
		doc["excerpt"] = Excerpt(field_val, max_char)
	}
}

type PagingInfo struct {
	Result 		[]paging.Pelem
	Skip, Current_page, All_results, Limit	int
	Paramkey, Url	string
}

// png = path and query
// In the CMS you can access it from uni.P + "?" + uni.Req.URL.RawQuery.
func DoPaging(db *mgo.Database, collection string, query map[string]interface{}, page_num_key string, get map[string][]string, pnq string, limit int) PagingInfo {
	var current_page int
	num_str, has := get[page_num_key]
	if !has {
		current_page = 1
	} else {
		val, err := strconv.ParseInt(num_str[0], 10, 32)
		if err == nil {
			current_page = int(val)
		} else {
			current_page = 1
		}
	}
	all_results, _ := db.C(collection).Find(query).Count()		// TODO: think about the error here.
	nav, _ := paging.P(current_page, all_results/limit + 1, 3, pnq)
	skip := (current_page - 1) * limit
	return PagingInfo{
		Result: 		nav,
		Skip: 			skip,
		Current_page: 	current_page,
		Limit:			limit,
		All_results:	all_results,
		Paramkey:		page_num_key,
		Url:			pnq,			
	}
}

// Convenience function.
func RunQuery(db *mgo.Database, query_name string, query map[string]interface{}, get map[string][]string, path_n_query string) map[string]interface{} {
	queries := map[string]interface{}{
		query_name : query,
	}
	return RunQueries(db, queries, get, path_n_query)
}

func toInt(num interface{}) int {
	switch val := num.(type) {
		case float64:
			return int(val)
		case int:
			return val
	}
	panic(fmt.Sprintf("Unkown type %T.", num))
}

// c: 		collection			string
// q: 		query				map[string]interface{}
// p:		page number key		string							This is used to extract the page nubver from get parameters. Also activates paging.	
//																Only works with limit.
// sk: 		skip				float64/int						Hardcoded value, barely useful (see p instead)
// l:		limit				float64/int
// so:		sort				string							Example: "-created"
//
// TODO: check for validity of type assertions.
func RunQueries(db *mgo.Database, queries map[string]interface{}, get map[string][]string, path_n_query string) map[string]interface{} {
	qs := make(map[string]interface{})
	for name, z := range queries {
		v := z.(map[string]interface{})
		_, coll_ok := v["c"]
		_, quer_ok := v["q"]
		if !coll_ok || !quer_ok { continue }
		q := db.C(v["c"].(string)).Find(v["q"])
		if skip, skok := v["sk"]; skok {
			q.Skip(toInt(skip))
		}
		if limit, lok := v["l"]; lok {
			q.Limit(toInt(limit))
		}
		if sort, sook := v["so"]; sook {
			if sort_string, is_str := sort.(string); is_str {
				q.Sort(sort_string)
			} else if sort_slice, is_sl := sort.([]interface{}); is_sl {
				q.Sort(jsonp.ToStringSlice(sort_slice)...)
			}
		}
		if p, pok := v["p"]; pok {
			if limit, lok := v["l"]; lok {	// Only makes sense with limit.
				paging_inf := DoPaging(db, v["c"].(string), v["q"].(map[string]interface{}), p.(string), get, path_n_query, toInt(limit))
				qs[name + "_navi"] = paging_inf
				q.Skip(paging_inf.Skip)
			}
		}
		var res []interface{}
		err := q.All(&res)
		if err != nil { qs[name] = err.Error(); continue }
		res = basic.Convert(res).([]interface{})
		if ex, ex_ok := v["ex"]; ex_ok {
			ex_m, ex_is_m := ex.(map[string]interface{})
			if ex_is_m && len(ex_m) == 1 { 
				CreateExcerpts(res, ex_m)
			}
		}
		dont_query := map[string]interface{}{"password":0}
		resolver.ResolveAll(db, res, dont_query)
		qs[name] = res
	}
	return qs
}

// Decides if a string should be localized.
const Min_loc_len = 8		// $loc.a.b
func IsLocString(s string) bool {
	return len(s) > Min_loc_len && string(s[0:4]) == "$loc." && strings.Index(s, ".") != -1
}

// Extracts the name of the localization file from the given loc string.
func ExtractLocName(s string) string {
	return strings.Split(s, ".")[1]
}

// TODO: This logic is very similar to what is being done in opesun/resolver. Check if a shared pattern could be extracted and reused.
func collect(i interface{}) []string {
	locfiles := []string{}
	switch val := i.(type) {
	case []interface{}:
		for _, v := range val {
			locfiles = append(locfiles, collect(v)...)
		}
	case map[string]interface{}:
		for _, v := range val {
			locfiles = append(locfiles, collect(v)...)
		}
	case string:
		if IsLocString(val) {
			locfiles = append(locfiles, ExtractLocName(val))
		}
	}
	return locfiles
}

// s absolute filepath
func locReader(s string) (map[string]interface{}, error) {
	file, err := ioutil.ReadFile(s)
	if err != nil { return nil, err }
	var v interface{}
	err = json.Unmarshal(file, &v)
	return v.(map[string]interface{}), err
}

// Extracts used multilingual variables from a template with regexp.
func CollectFromTempl(file_content string) map[string]struct{} {
	r := regexp.MustCompile(".loc.([a-zA-Z_.:/-])*")
	s := r.FindAllString(file_content, -1)
	c := map[string]struct{}{}
	for _, v := range s {
		spl := strings.Split(v, ".")
		if len(spl) > 3 {
			c[spl[2]] = struct{}{}
		}
	}
	return c
}

func CollectFromMap(dat map[string]interface{}) map[string]struct{} {
	sl := collect(dat)
	c := map[string]struct{}{}
	for _, v := range sl {
		c[v] = struct{}{}
	}
	return c
}

// Takes a list of localization filenames and tries to load every one of them, first from the template, then from the modules.
func ReadFiles(root, tplpath string, user_langs []string, locfiles map[string]struct{}, loc_reader func (s string) (map[string]interface{}, error)) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for i, _ := range locfiles {
		for _, lang := range user_langs {
			path := filepath.Join(root, tplpath, "loc", i + "." + lang)
			ma, err := loc_reader(path)
			if err == nil {
				ret[i] = ma
				break
			}
			path = filepath.Join(root, "modules", i, "tpl/loc", lang + ".json")
			ma, err = loc_reader(path)
			if err == nil {
				ret[i] = ma
				break
			}
		}
	}
	return ret, nil
}

// tplpath is public/default or private/127.0.0.1/default
func LoadLocStrings(dat map[string]interface{}, user_langs []string, root, tplpath string, loc_reader func (s string) (map[string]interface{}, error)) (map[string]interface{}, error) {
	if loc_reader == nil {
		loc_reader = locReader
	}
	locfiles := CollectFromMap(dat)
	return ReadFiles(root, tplpath, user_langs, locfiles, loc_reader)
}

// tplpath is public/default or private/127.0.0.1/default
func LoadLocTempl(file_content string, user_langs []string, root, tplpath string, loc_reader func (s string) (map[string]interface{}, error)) (map[string]interface{}, error) {
	if loc_reader == nil {
		loc_reader = locReader
	}
	locfiles := CollectFromTempl(file_content)
	return ReadFiles(root, tplpath, user_langs, locfiles, loc_reader)
}
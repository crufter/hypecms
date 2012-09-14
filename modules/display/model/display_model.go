package display_model

import (
	"fmt"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/jsonp"
	"github.com/opesun/paging"
	"github.com/opesun/resolver"
	"labix.org/v2/mgo"
	"strconv"
	"strings"
)

// Cuts a long string at max_char_count, taking a word boundary into account.
func Excerpt(s string, max_char_count int) string {
	if len(s) < max_char_count {
		return s
	}
	ind := strings.LastIndex(s[:max_char_count], " ")
	if ind == -1 {
		return s
	}
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
		if !ok {
			continue
		}
		doc["excerpt"] = Excerpt(field_val, max_char)
	}
}

type PagingInfo struct {
	Result                                 []paging.Pelem
	Skip, Current_page, All_results, Limit int
	Paramkey, Url                          string
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
	all_results, _ := db.C(collection).Find(query).Count() // TODO: think about the error here.
	nav, _ := paging.P(current_page, all_results/limit+1, 3, pnq)
	skip := (current_page - 1) * limit
	return PagingInfo{
		Result:       nav,
		Skip:         skip,
		Current_page: current_page,
		Limit:        limit,
		All_results:  all_results,
		Paramkey:     page_num_key,
		Url:          pnq,
	}
}

// Convenience function.
func RunQuery(db *mgo.Database, query_name string, query map[string]interface{}, get map[string][]string, path_n_query string) map[string]interface{} {
	queries := map[string]interface{}{
		query_name: query,
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
		if !coll_ok || !quer_ok {
			continue
		}
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
			if limit, lok := v["l"]; lok { // Only makes sense with limit.
				paging_inf := DoPaging(db, v["c"].(string), v["q"].(map[string]interface{}), p.(string), get, path_n_query, toInt(limit))
				qs[name+"_navi"] = paging_inf
				q.Skip(paging_inf.Skip)
			}
		}
		var res []interface{}
		err := q.All(&res)
		if err != nil {
			qs[name] = err.Error()
			continue
		}
		res = basic.Convert(res).([]interface{})
		if ex, ex_ok := v["ex"]; ex_ok {
			ex_m, ex_is_m := ex.(map[string]interface{})
			if ex_is_m && len(ex_m) == 1 {
				CreateExcerpts(res, ex_m)
			}
		}
		var resolve_fields map[string]interface{}
		if val, has := v["r"]; has {
			resolve_fields = val.(map[string]interface{})
		} else {
			resolve_fields = map[string]interface{}{"password": 0}
		}
		resolver.ResolveAll(db, res, resolve_fields)
		qs[name] = res
	}
	return qs
}

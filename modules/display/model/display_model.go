package display_model

import(
	"labix.org/v2/mgo"
	"github.com/opesun/jsonp"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/paging"
	"strconv"
)

// png = path and query
// In the cms you can access it from uni.P + "?" + uni.Req.URL.RawQuery.
func DoPaging(db *mgo.Database, collection string, query map[string]interface{}, page_num_key string, get map[string][]string, pnq string, limit int) (int, []paging.Pelem) {
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
	max_results, _ := db.C(collection).Find(query).Count()		// TODO: think about the error here.
	nav, _ := paging.P(current_page, max_results/limit + 1, 3, pnq)
	return (current_page - 1) * limit, nav
}

// c: 		collection			string
// q: 		query				map[string]interface{}
// p:		page number key		string							This is used to extract the page nubver from get parameters. Also activates paging.	
//																Only works with limit.
// sk: 		skip				float64 (int in fact)			Hardcoded value, barely useful (see p instead)
// l:		limit				float64 (int in fact)
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
			q.Skip(int(skip.(float64)))
		}
		if limit, lok := v["l"]; lok {
			q.Limit(int(limit.(float64)))
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
				skip_amount, navigation := DoPaging(db, v["c"].(string), v["q"].(map[string]interface{}), p.(string), get, path_n_query, int(limit.(float64)))
				qs[name + "_navi"] = navigation
				q.Skip(skip_amount)
			}
		}
		var res []interface{}
		err := q.All(&res)
		if err != nil { qs[name] = err.Error() }
		qs[name] = res
	}
	for i, _ := range qs {
		// Can be []pagin.Pelem too.
		if _, is_islice := qs[i].([]interface{}); is_islice {
			qs[i] = basic.Convert(qs[i]).([]interface{})
			basic.IdsToStrings(qs[i])
		}
	}
	return qs
}
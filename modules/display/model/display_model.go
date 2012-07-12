package display_model

import(
	"labix.org/v2/mgo"
	"github.com/opesun/jsonp"
	"github.com/opesun/hypecms/model/basic"
)

// n:		name			string
// c: 		collection		string
// q: 		query			map[string]interface{}
// sk: 		skip			float64 (int in fact)
// l:		limit			float64 (int in fact)
//
// TODO: check for validity of type assertions.
func RunQueries(db *mgo.Database, queries []interface{}) map[string]interface{} {
	qs := make(map[string]interface{})
	for _, z := range queries {
		v := z.(map[string]interface{})
		q := db.C(v["c"].(string)).Find(v["q"])
		if skip, skok := v["sk"]; skok {
			q.Skip(int(skip.(float64)))
		}
		if limit, lok := v["l"]; lok {
			q.Limit(int(limit.(float64)))
		}
		if sort, sook := v["so"]; sook {
			q.Sort(jsonp.ToStringSlice(sort)...)
		}
		var res []interface{}
		q.All(&res)
		qs[v["n"].(string)] = res
	}
	for i, _ := range qs {
		qs[i] = basic.Convert(qs[i]).([]interface{})
		basic.IdsToStrings(qs[i])
	}
	return qs
}
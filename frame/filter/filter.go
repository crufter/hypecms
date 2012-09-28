package filter

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	//"net/url"
)

type filterMod struct {
	skip		int
	limit		int
	page		int
	perPage		int
}

type Filter struct {
	db 			*mgo.Database
	coll 		string
	filterMod	filterMod
	parents		map[string][]bson.ObjectId
	query		map[string]interface{}
}

func ToQuery(a url.Values) map[string]interface{} {
	r := map[string]interface{}{}
	for i, v := range a {
		var vi []interface{}
		for j, x := range v {
			var val interface{}
			switch i {
			case "id":
				val = bson.ObjectIdHex(x)
			default:
				val = x
			}
			vi = append(vi, val)
		}
		if len(vi) > 1 {
			r[i] = map[string]interface{}{
				"$in": vi,
			}
		} else {
			r[i] = vi[0]
		}
	}
}

func New(db *mgo.Database, coll string, parents map[string][]bson.ObjectId, query map[string]interface{}) *Filter {
	return &Filter{db, coll, filterMod{}, parents, query}
}

func merge(q map[string]interface{}, p map[string][]bson.ObjectId) map[string]interface{} {
	r := map[string]interface{}{}
	for i, v := range q {
		r[i] = v
	}
	for i, v := range p {
		r[i] = map[string]interface{}{
			"$in": v,
		}
	}
	return r
}

func (f *Filter) Find() ([]interface{}, error) {
	var res []interface{}
	q := merge(f.query, f.parents)
	err := f.db.C(f.coll).Find(q).All(&res)
	return res, err
}

func (f *Filter) Insert(d map[string]interface{}) error {
	i := merge(d, f.parents)
	return f.db.C(f.coll).Insert(i)
}

func (f *Filter) Update(upd_query map[string]interface{}) error {
	q := merge(f.query, f.parents)
	_, err := f.db.C(f.coll).UpdateAll(q, upd_query)
	return err
}

func (f *Filter) Ids() ([]bson.ObjectId, error) {
	if val, has := f.query["id"]; has && len(f.query) == 1 && len(f.parents) == 1 {
		ids := val.(map[string]interface{})["$in"].([]interface{})
		ret := []bson.ObjectId{}
		for _, v := range ids {
			ret = append(ret, v.(bson.ObjectId))
		}
		return ret, nil
	}
	q := merge(f.query, f.parents)
	var res []interface{}
	err := f.db.C(f.coll).Find(q).All(&res)
	if err != nil {
		return nil, err
	}
	ret := []bson.ObjectId{}
	for _, v := range res {
		ret = append(ret, v.(bson.M)["_id"].(bson.ObjectId))
	}
	return ret, nil
}

func (f *Filter) Remove() error {
	return f.db.C(f.coll).Remove(f.query)
}
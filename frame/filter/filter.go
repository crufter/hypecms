package filter

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/url"
	"fmt"
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

func (f *Filter) Visualize() {
	fmt.Println("<<<")
	fmt.Println("db", f.db)
	fmt.Println("coll", f.coll)
	fmt.Println("fmod", f.filterMod)
	fmt.Println("parents", f.parents)
	fmt.Println("query", f.query)
	fmt.Println(">>>")
}

func Reduce(a ...*Filter) (*Filter, error) {
	l := len(a)
	if l == 0 {
		return &Filter{}, fmt.Errorf("Nothing to reduce.")
	}
	if l == 1 {
		return a[0], nil
	}
	prev := a[0]
	for i:=1;i<l;i++ {
		ids, err := prev.Ids()
		if err != nil {
			return &Filter{}, err
		}
		a[i].SetParents("_"+prev.Coll(), ids)
		prev = a[i]
	}
	return prev, nil
}

func ToData(a url.Values) map[string]interface{} {
	r := map[string]interface{}{}
	for i, v := range a {
		vi := []interface{}{}
		for _, x := range v {
			vi = append(vi, x)
		}
		r[i] = vi
	}
	return r
}

func ToQuery(a url.Values) map[string]interface{} {
	r := map[string]interface{}{}
	for i, v := range a {
		var vi []interface{}
		for _, x := range v {
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
	return r
}

func New(db *mgo.Database, coll string, query map[string]interface{}) *Filter {
	return &Filter{db, coll, filterMod{}, map[string][]bson.ObjectId{}, query}
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
	f.Visualize()
	fmt.Println(q)
	err := f.db.C(f.coll).Find(q).All(&res)
	fmt.Println(res)
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

func (f *Filter) Coll() string {
	return f.coll
}

func (f *Filter) SetParents(fieldname string, a []bson.ObjectId) {
	f.parents[fieldname] = a
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
package content_model

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/resolver"
	"github.com/opesun/extract"
	"github.com/opesun/jsonp"
	"github.com/opesun/slugify"
	"github.com/opesun/hypecms/model/basic"
	"fmt"
	"strings"
)

// Precedence: type && op, type, op, all
func requiredLevel(content_options map[string]interface{}, typ, op string) int {
	type_op_lev, has := jsonp.Get(content_options, "types." + typ + "." + op + "_level")
	if has {
		return int(type_op_lev.(float64))
	}
	type_lev, has := jsonp.Get(content_options, "types." + typ + ".level")
	if has {
		return int(type_lev.(float64))
	}
	lev, has := content_options["level"]
	if has {
		return int(lev.(float64))
	}
	return 300
}

// op is insert/update/delete
func AllowsContent(db *mgo.Database, inp map[string][]string, content_options map[string]interface{}, user_id bson.ObjectId, user_level int, op string) error {
	rule := map[string]interface{}{
		"type": "must",
		"id": "must",
	}
	dat, ex_err := extract.New(rule).Extract(inp)
	if ex_err != nil { return ex_err }
	var inserting bool
	if len(dat["id"].(string)) == 0 {
		inserting = true
	}
	req_lev := requiredLevel(content_options, dat["type"].(string), op)
	if req_lev > user_level {
		return fmt.Errorf("You have no rights to manage contents.")
	}
	if user_level < 200 && !inserting {
		auth, find_err := findContentAuthor(db, basic.StripId(dat["_id"].(string)))
		if find_err != nil { return find_err }
		if auth.Hex() != user_id.Hex() {
			return fmt.Errorf("You are not the rightous owner of the content.")
		}
	}
	return nil
}

// returns nil if not found
func find(db *mgo.Database, content_id string) map[string]interface{} {
	if len(content_id) != 24 { return nil }
	q := bson.M{
		"_id": bson.ObjectIdHex(content_id),
	}
	var v interface{}
	err := db.C("contents").Find(q).One(&v)
	if err != nil { return nil }
	if v == nil { return nil }
	return basic.Convert(v).(map[string]interface{})
}

func findContentAuthor(db *mgo.Database, content_id string) (bson.ObjectId, error) {
	cont := find(db, content_id)
	if cont == nil {
		return "", fmt.Errorf("Content not found.")
	}
	auth, has := cont["created_by"]
	if !has {
		return "", fmt.Errorf("Content has no author.")
	}
	return auth.(bson.ObjectId), nil
}

// TODO: rethink. TODO: Add number support.
// Walks an entire JSON tree recursively, and converts everything to string it can find.
func walkDeep(i interface{}) []string {
	switch val := i.(type) {
	case map[string]interface{}:
		ret := []string{}
		for _, v := range val {
			ret = append(ret, walkDeep(v)...)
		}
		return ret
	case []interface{}:
		ret := []string{}
		for _, v := range val {
			ret = append(ret, walkDeep(v)...)
		}
		return ret
	case string:
		return []string{val}
	case bson.M:
		panic("This should definitely not happen, basic.Convert is no good.")
	}
	return []string{}
}

func filterDupes(s []string) []string{
	ret := []string{}
	c := map[string]struct{}{}
	for _, v := range s {
		if _, has := c[v]; !has {
			c[v] = struct{}{}
			ret = append(ret, v)
		}
	}
	return ret
}

func filterTooShort(s []string, min_len int) []string{
	ret := []string{}
	for _, v := range s {
		if len(v) >= min_len {
			ret = append(ret, v)
		}
	}
	return ret
}

// Mostly for in-developement use.
func RegenerateFulltext(db *mgo.Database) error {
	return nil
}

// TODO:
// Resolve may be extended so you can set the queried fields. Now the whole object is queried, and it can cause problems, like
// including the password of the author and such.
func generateFulltext(db *mgo.Database, id bson.ObjectId) []string {
	var res interface{}
	db.C("contents").Find(m{"_id": id}).One(&res)
	dat := basic.Convert(res).(map[string]interface{})
	resolver.ResolveOne(db, dat)
	dat = basic.Convert(dat).(map[string]interface{})
	non_split := walkDeep(dat)
	split := []string{}
	for _, v := range non_split {
		split = append(split, strings.Split(v, " ")...)
	}
	slugified := []string{}
	for _, v := range split {
		slugified = append(slugified, strings.Trim(slugify.S(v), ",.:;"))
	}
	slugified = filterDupes(slugified)
	return filterTooShort(slugified, 3)
}

func saveFulltext(db *mgo.Database, id bson.ObjectId) error {
	fulltext := generateFulltext(db, id)
	return db.C("contents").Update(m{"_id": id}, m{"$set": m{"fulltext": fulltext}})
}

func GenerateKeywords(s string) []string {
	split := strings.Split(s, " ")
	slugified := []string{}
	for _, v := range split {
		slugified = append(slugified, strings.Trim(slugify.S(v), ",.:;"))
	}
	return slugified
}

// Generates [{"fulltext": \^keyword1\}, {"fulltext": \^keyword2\}]
// With this query we can create a good enough full text search, which can search at the beginning of the keywords.
// We could write regexes which searches in the middle of the words too, but that query could not uzilize the btree indexes of mongodb.
// This solution must be efficient, assuming mongodb does the expected sane things: utilizing indexes with ^ regexes, "$and" queries and arrays.
func GenerateQuery(s string) []interface{} {
	sl := GenerateKeywords(s)
	and := []interface{}{}
	for _, v := range sl {
		and = append(and, map[string]interface{}{"fulltext": bson.RegEx{Pattern: "^" + v}})
	}
	return and
}

func Insert(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId) (bson.ObjectId, error) {
	id, hasid := dat["id"]
	if hasid && len(id[0]) > 0 {
		return "", fmt.Errorf("Can't insert an object wich already has an id.")
	}
	typ, hastype := dat["type"]
	if !hastype {
		return "", fmt.Errorf("No type when inserting content.")
	}
	ins_dat, extr_err := extract.New(rule).Extract(dat)
	if extr_err != nil {
		return "", extr_err
	}
	basic.DateAndAuthor(rule, ins_dat, user_id)
	ins_dat["type"] = typ[0]
	_, has_tags := ins_dat[Tag_fieldname_displayed]
	if has_tags {
		addTags(db, ins_dat, "", "insert")
	}
	err := basic.Inud(db, ev, ins_dat, "contents", "insert", "")
	if err != nil { return "", err }
	ret_id := ins_dat["_id"].(bson.ObjectId)
	_, has_fulltext := rule["fulltext"]
	if has_fulltext {
		saveFulltext(db, ret_id)
	}
	return ret_id, nil
}

func Update(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId) error {
	id, hasid := dat["id"]
	if !hasid {
		return fmt.Errorf("No id when updating content.")
	}
	typ, hastype := dat["type"]
	if !hastype {
		return fmt.Errorf("No type when updating content.")
	}
	upd_dat, extr_err := extract.New(rule).Extract(dat)
	if extr_err != nil {
		return extr_err
	}
	basic.DateAndAuthor(rule, upd_dat, user_id)
	upd_dat["type"] = typ[0]
	_, has_tags := upd_dat[Tag_fieldname_displayed]
	if has_tags {
		addTags(db, upd_dat, id[0], "update")
	}
	ret_err := basic.Inud(db, ev, upd_dat, "contents", "update", id[0])
	_, has_fulltext := rule["fulltext"]
	if has_fulltext {
		id_bson := bson.ObjectIdHex(basic.StripId(id[0]))
		saveFulltext(db, id_bson)
	}
	return ret_err
}

func Delete(db *mgo.Database, ev ifaces.Event, id []string, user_id bson.ObjectId) []error {
	var errs []error
	for _, v := range id {
		errs = append(errs, basic.Inud(db, ev, nil, "contents", "delete", v))
	}
	return errs
}

// Called from Front hook.
// Find slug value by given key.
func FindContent(db *mgo.Database, keys []string, val string) (map[string]interface{}, bool) {
	query := bson.M{}
	if len(keys) == 0 {
		return nil, false
	} else if len(keys) == 1 {
		if keys[0] == "_id" && len(val) == 24 {			// TODO: check for validity of id.
			query[keys[0]] = bson.ObjectIdHex(val)
		} else {
			query[keys[0]] = val
		}
	} else {
		or := []map[string]interface{}{}
		for _, v := range keys {
			if v == "_id" && len(v) == 24 {				// TODO: check fir validity of id.
				or = append(or, map[string]interface{}{v: bson.ObjectIdHex(val)})
			} else {
				or = append(or, map[string]interface{}{v: val})
			}
		}
		query["$or"] = or
	}
	var v interface{}
	db.C("contents").Find(query).One(&v)
	if v == nil {
		return nil, false
	}
	return basic.Convert(v).(map[string]interface{}), true
}

func SaveTypeConfig(db *mgo.Database, inp map[string][]string) error {
	rule := map[string]interface{}{
		"type": 		"must",
		"safe_delete":	"must",
	}
	dat, err := extract.New(rule).Extract(inp)
	fmt.Println(dat)
	if err != nil { return err }
	// TODO: finish.
	return nil
}

func SavePersonalTypeConfig(db *mgo.Database, inp map[string][]string, user_id bson.ObjectId) error {
	return nil
}

func Install(db *mgo.Database, id bson.ObjectId) error {
	content_options := m{
		"types": m {
			"blog": m{
				"rules" : m{
					"title": 1, "slug":1, "content": 1, Tag_fieldname_displayed: 1, "fulltext": false, basic.Created: false, basic.Created_by:false,
				},
			},
		},
	}
	q := m{"_id": id}
	upd := m{
		"$addToSet": m{
			"Hooks.Front": "content",
		},
		"$set": m{
			"Modules.content": content_options,
			"Display-points.index.queries.blog": m{
				"q":	m{ "type": "blog"},
				"so":	"-created",
				"p":	"page",
			},
		},
	}
	return db.C("options").Update(q, upd)
}

func Uninstall(db *mgo.Database, id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$pull": m{
			"Hooks.Front": "content",
		},
		"$unset": m{
			"Modules.content": 1,
			"Display-points.index.queries.blog": 1,
		},
	}
	return db.C("options").Update(q, upd)
}
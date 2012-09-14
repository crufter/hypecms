package content_model

import (
	"fmt"
	"github.com/opesun/extract"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/resolver"
	"github.com/opesun/slugify"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

const (
	Cname = "contents"
	not_impl = "Not implemented yet."
)

// Checks if one is entitled to modify other peoples content.
func CanModifyContent(db *mgo.Database, inp map[string][]string, correction_level int, user_id bson.ObjectId, user_level int) error {
	rule := map[string]interface{}{
		"id":   "must",
	}
	dat, err := extract.New(rule).Extract(inp)
	if err != nil {
		return err
	}
	if user_level < correction_level {
		content := find(db, dat["id"].(string))
		if content == nil {
			return fmt.Errorf("Can't find content.")
		}
		auth, err := contentAuthor(content)
		if err != nil {
			return err
		}
		if auth.Hex() != user_id.Hex() {
			return fmt.Errorf("You can not modify this type of content if it is not yours.")
		}
	}
	return nil
}

// returns nil if not found
func find(db *mgo.Database, content_id string) map[string]interface{} {
	content_bsonid := basic.ToIdWithCare(content_id)
	q := bson.M{
		"_id": content_bsonid,
	}
	var v interface{}
	err := db.C("contents").Find(q).One(&v)
	if err != nil {
		return nil
	}
	return basic.Convert(v).(map[string]interface{})
}

// Returns the type of a given content.
func TypeOf(db *mgo.Database, content_id bson.ObjectId) (string, error) {
	var v interface{}
	q := m{"_id": content_id}
	err := db.C("contents").Find(q).One(&v)
	if err != nil {
		return "", err
	}
	con := basic.Convert(v.(bson.M)).(map[string]interface{})
	typ, has := con["type"]
	if !has {
		return "", fmt.Errorf("Content is malformed: has no type.")
	}
	return typ.(string), nil
}

// Unused atm. Probably will be deleted.
func typed(db *mgo.Database, content_id bson.ObjectId, typ string) bool {
	var v interface{}
	q := m{"_id": content_id, "type": typ}
	err := db.C("contents").Find(q).One(&v)
	return err == nil // One problem is we can't differentiate between not found and IO error here...
}

func contentAuthor(content map[string]interface{}) (bson.ObjectId, error) {
	auth, has := content["created_by"]
	if !has {
		return "", fmt.Errorf("Content has no author.")
	}
	return auth.(bson.ObjectId), nil
}

// Not used.
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

func filterDupes(s []string) []string {
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

func filterTooShort(s []string, min_len int) []string {
	ret := []string{}
	for _, v := range s {
		if len(v) >= min_len {
			ret = append(ret, v)
		}
	}
	return ret
}

// Very simple way of building a fulltext field - split it by spaces, then slug each string.
func simpleFulltext(non_split []string) []string {
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

// Generates the fulltext field by querying the content, resolving it's foreign keys, then trimming them
func generateFulltext(db *mgo.Database, id bson.ObjectId) []string {
	var res interface{}
	db.C("contents").Find(m{"_id": id}).One(&res)
	dat := basic.Convert(res).(map[string]interface{})
	fields := map[string]interface{}{
		"name":  1,
		"slug":  1,
		"title": 1,
	}
	resolver.ResolveOne(db, dat, fields)
	dat = basic.Convert(dat).(map[string]interface{})
	non_split := walkDeep(dat)
	return simpleFulltext(non_split)
}

// Queries a content and then updates it to save the fulltext field.
func saveFulltext(db *mgo.Database, id bson.ObjectId) error {
	fulltext := generateFulltext(db, id)
	return db.C("contents").Update(m{"_id": id}, m{"$set": m{"fulltext": fulltext}})
}

// Mostly for in-development use. Iterates trough all contents in the contents collection and regenerates their fulltext field.
func RegenerateFulltext(db *mgo.Database) error {
	return fmt.Errorf(not_impl)
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

func mergeMaps(a, b map[string]interface{}) {
	if a == nil || b == nil {
		return
	}
	for i, v := range b {
		_, has := a[i]
		if has {
			panic("Overwriting existing value.")
		}
		a[i] = v
	}
}

func Insert(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId) (bson.ObjectId, error) {
	return insert(db, ev, rule, dat, user_id, nil)
}

func InsertWithFix(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId, fixvals map[string]interface{}) (bson.ObjectId, error) {
	return insert(db, ev, rule, dat, user_id, fixvals)
}

func insert(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId, fixvals map[string]interface{}) (bson.ObjectId, error) {
	// Could check for id here, alert if we found one.
	rule["type"] = "must"
	rule["draft_id"] = "must" // Can be draft, or version.
	ins_dat, extr_err := extract.New(rule).Extract(dat)
	if extr_err != nil {
		return "", extr_err
	}
	typ := ins_dat["type"].(string)
	basic.DateAndAuthor(rule, ins_dat, user_id, false)
	_, has_tags := ins_dat[Tag_fieldname_displayed]
	if has_tags {
		addTags(db, ins_dat, "", "insert", typ)
	}
	basic.Slug(rule, ins_dat)
	mergeMaps(ins_dat, fixvals)
	err := basic.InudVersion(db, ev, ins_dat, "contents", "insert", "")
	if err != nil {
		return "", err
	}
	ret_id := ins_dat["_id"].(bson.ObjectId)
	_, has_fulltext := rule["fulltext"]
	if has_fulltext {
		saveFulltext(db, ret_id)
	}
	return ret_id, nil
}

func Update(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId) error {
	return update(db, ev, rule, dat, user_id, nil)
}

func UpdateWithFix(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId, fixvals map[string]interface{}) error {
	return update(db, ev, rule, dat, user_id, fixvals)
}

func update(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string, user_id bson.ObjectId, fixvals map[string]interface{}) error {
	rule["id"] = "must"
	rule["type"] = "must"
	rule["draft_id"] = "must"
	upd_dat, extr_err := extract.New(rule).Extract(dat)
	if extr_err != nil {
		return extr_err
	}
	id := upd_dat["id"].(string)
	typ := upd_dat["type"].(string)
	basic.DateAndAuthor(rule, upd_dat, user_id, true)
	upd_dat["type"] = typ
	_, has_tags := upd_dat[Tag_fieldname_displayed]
	if has_tags {
		addTags(db, upd_dat, id, "update", typ)
	}
	basic.Slug(rule, upd_dat)
	mergeMaps(upd_dat, fixvals)
	err := basic.InudVersion(db, ev, upd_dat, Cname, "update", id)
	if err != nil {
		return err
	}
	_, has_fulltext := rule["fulltext"]
	id_bson := bson.ObjectIdHex(basic.StripId(id))
	if has_fulltext {
		saveFulltext(db, id_bson)
	}
	return nil
}

func Delete(db *mgo.Database, ev ifaces.Event, id []string, user_id bson.ObjectId) []error {
	var errs []error
	for _, v := range id {
		errs = append(errs, basic.Inud(db, ev, nil, Cname, "delete", v))
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
		if keys[0] == "_id" && len(val) == 24 { // TODO: check for validity of id.
			query[keys[0]] = bson.ObjectIdHex(val)
		} else {
			query[keys[0]] = val
		}
	} else {
		or := []map[string]interface{}{}
		for _, v := range keys {
			if v == "_id" && len(v) == 24 { // TODO: check fir validity of id.
				or = append(or, map[string]interface{}{v: bson.ObjectIdHex(val)})
			} else {
				or = append(or, map[string]interface{}{v: val})
			}
		}
		query["$or"] = or
	}
	var v interface{}
	db.C(Cname).Find(query).One(&v)
	if v == nil {
		return nil, false
	}
	return basic.Convert(v).(map[string]interface{}), true
}

func SaveTypeConfig(db *mgo.Database, inp map[string][]string) error {
	rule := map[string]interface{}{
		"type":        "must",
		"safe_delete": "must",
	}
	_, err := extract.New(rule).Extract(inp) // _ = dat
	if err != nil {
		return err
	}
	// TODO: finish.
	return nil
}

func SavePersonalTypeConfig(db *mgo.Database, inp map[string][]string, user_id bson.ObjectId) error {
	return nil
}

func Install(db *mgo.Database, id bson.ObjectId) error {
	content_options := m{
		"actions": m{
			"insert_comment": m{
				"auth": false,
			},
		},
		"types": m{
			"blog": m{
				"actions": m{
					"insert_comment": m{
						"auth": m{
							"min_lev": 0,
							"no_puzzles_lev": 2,
							"hot_reg": 2,
						},
					},
				},
				"comment_rules": m{
					basic.Created:     false,
					basic.Created_by:  false,
					"comment_content": 1,
				},
				"rules": m{
					"title":                 1,
					"slug":                  1,
					"content":               1,
					Tag_fieldname_displayed: 1,
					"fulltext":              false,
					basic.Created:           false,
					basic.Created_by:        false,
					basic.Last_modified:     false,
					basic.Last_modified_by:  false,
				},
				"non_versioned_fields": m{
					"comments": 1,
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
				"ex": map[string]interface{}{
				"content": 300,
				},
				"c":  Cname,
				"l":  10,
				"q":  m{"type": "blog"},
				"so": "-created",
				"p":  "page",
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
			"Modules.content":                   1,
			"Display-points.index.queries.blog": 1,
		},
	}
	return db.C("options").Update(q, upd)
}

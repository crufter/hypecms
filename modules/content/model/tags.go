package content_model

import (
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/hypecms/model/patterns"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/slugify"
	"strings"
	"fmt"
)

const(
	Tag_fieldname 				= "_tags"					// goes to database
	Tag_fieldname_displayed 	= "tags"		// comes from user interface, it is in the rules
	Tag_cname 					= "tags"
	Count_fieldname 			= "count"
)

type m map[string]interface{}

func createM(slugs []string) map[string]struct{} {
	ret := map[string]struct{}{}
	for _, v := range slugs {
		ret[v] = struct{}{}
	}
	return ret
}

func mToSSlice(ma map[string]struct{}) []string {
	ret := []string{}
	for i, _ := range ma {
		ret = append(ret, i)
	}
	return ret
}

func separateTags(db *mgo.Database, tag_sl []string) ([]bson.ObjectId, []string) {
	var i []interface{}
	slug_sl := []string{}
	for _, val := range tag_sl {
		slug := slugify.S(val)
		slug_sl = append(slug_sl, slug)
	}
	db.C(Tag_cname).Find(m{"slug":m{ "$in":slug_sl}}).Limit(0).All(&i)
	ret_ids := []bson.ObjectId{}
	contains := createM(slug_sl)
	i = basic.Convert(i).([]interface{})
	for _, v := range i {
		val := v.(map[string]interface{})
		ret_ids = append(ret_ids, val["_id"].(bson.ObjectId))
		delete(contains, val["slug"].(string))
	}
	return ret_ids, mToSSlice(contains)
}

func inc(db *mgo.Database, ids []bson.ObjectId) error {
	return patterns.IncAll(db, Tag_cname, ids, Count_fieldname, 1)
}

func dec(db *mgo.Database, ids []bson.ObjectId) error {
	return patterns.IncAll(db, Tag_cname, ids, Count_fieldname, -1)
}

func insertTags(db *mgo.Database, tag_sl []string) []bson.ObjectId {
	ret := []bson.ObjectId{}
	for _, v := range tag_sl {
		if len(v) == 0 { continue }
		id := bson.NewObjectId()
		slug := slugify.S(v)
		tag := m{"_id": id, "slug":slug, "name":v, Count_fieldname:0}
		db.C(Tag_cname).Insert(tag)
		ret = append(ret, id)
	}
	return ret
}

func createMObjectId(a []bson.ObjectId) map[bson.ObjectId]struct{} {
	ret := map[bson.ObjectId]struct{}{}
	for _, v := range a {
		ret[v] = struct{}{}
	}
	return ret
}

// Gives back what a contains but b not.
func diffIds(a []bson.ObjectId, b []bson.ObjectId) []bson.ObjectId {
	b_cache := createMObjectId(b)
	ret := []bson.ObjectId{}
	for _, v := range a {
		if _, has := b_cache[v]; !has {
			ret = append(ret, v)
		}
	}
	return ret
}

// "a, b, c" ... "a, c" => decrement b, old node.js version logic
func handleCount(db *mgo.Database, old_ids []bson.ObjectId, new_ids []bson.ObjectId) {
	dec_ids := diffIds(old_ids, new_ids)
	inc_ids := diffIds(new_ids, old_ids)
	if len(inc_ids) > 0 {
		inc(db, inc_ids)
	}
	if len(dec_ids) > 0 {
		dec(db, dec_ids)
	}
}

func mergeIds(a []bson.ObjectId, b []bson.ObjectId) []bson.ObjectId {
	ret := []bson.ObjectId{}
	ret = append(ret, a...)
	ret = append(ret, b...)
	return ret
}

func toIdSlice(i []interface{}) []bson.ObjectId {
		ret := []bson.ObjectId{}
		for _, v := range i {
			ret = append(ret, v.(bson.ObjectId))
		}
		return ret
}

func addToSet(db *mgo.Database, content_id bson.ObjectId, tag_ids []bson.ObjectId) {
	q := m{"_id": content_id}
	upd := m{"$addToSet": m{Tag_fieldname: m{"$each": tag_ids}}}
	db.C(Cname).Update(q, upd)
}

func stripEmpty(sl []string) []string {
	ret := []string{}
	for _, v := range sl {
		if len(v) > 0 {
			ret = append(ret, v)
		}
	}
	return ret
}

// Creates nonexisting tags if needed and pushes the tag ids into the content, and increments the tag counters.
func addTags(db *mgo.Database, dat map[string]interface{}, id string, mod string) {
	content := map[string]interface{}{}
	if mod != "insert" {
		content = find(db, basic.StripId(id))
	}
	tags_i, _ := dat[Tag_fieldname_displayed]
	delete(dat, Tag_fieldname_displayed)
	tags := tags_i.(string)					// Example: "Cars, Bicycles"
	tags_sl := strings.Split(tags, ",")
	tags_sl = stripEmpty(tags_sl)
	switch mod {
		case "insert":
			existing_ids, to_insert_slugs := separateTags(db, tags_sl)
			inserted_ids := insertTags(db, to_insert_slugs)
			all_ids := mergeIds(existing_ids, inserted_ids)
			inc(db, all_ids)
			dat[Tag_fieldname] = all_ids
		case "update":
			existing_ids, to_insert_slugs := separateTags(db, tags_sl)
			inserted_ids := insertTags(db, to_insert_slugs)
			old_ids := toIdSlice(content[Tag_fieldname].([]interface{}))
			new_ids := mergeIds(existing_ids, inserted_ids)
			inc_ids := diffIds(new_ids, old_ids)
			inc(db, inc_ids)
			addToSet(db, content["_id"].(bson.ObjectId), new_ids)
		default:
			panic("Bad mode at addTags.")
	}
}

// Pulls one or more tag ids from a content and decrements the tag counters.
// UI can send "ObjectIdHex(\"abababab656b5a6b5a6b5a6b5\")" instead of "abababab656b5a6b5a6b5a6b5"
func PullTags(db *mgo.Database, content_id string, tag_ids []string) error {
	content_id = basic.StripId(content_id)
	content := find(db, content_id)
	if content == nil { return fmt.Errorf("Cant find content when pulling tags.") }
	tag_objectids := content[Tag_fieldname].([]interface{})
	cache := createMObjectId(toIdSlice(tag_objectids))
	to_pull := []bson.ObjectId{}
	for _, v := range tag_ids {
		tag_id := basic.StripId(v)
		tag_objectid := bson.ObjectIdHex(tag_id)
		if _, has := cache[tag_objectid]; has {
			to_pull = append(to_pull, tag_objectid)
		}
	}
	dec(db, to_pull)
	q := m{"_id": content["_id"].(bson.ObjectId) }
	upd := m{"$pullAll": m{Tag_fieldname: to_pull}}
	return db.C(Cname).Update(q, upd)
}

// Deletes a tag entirely.
func DeleteTag(db *mgo.Database, tag_id string) error {
	err := patterns.DeleteById(db, Tag_cname, tag_id)
	if err != nil {return err}
	return PullTagFromAll(db, tag_id)
}

func PullTagFromAll(db *mgo.Database, tag_id string) error {
	return patterns.PullFromAll(db, Tag_cname, Tag_fieldname, patterns.ToIdWithCare(tag_id))
}

func ListContentsByTag(db *mgo.Database, field string, value interface{}) ([]interface{}, error) {
	return patterns.FindParentAndChildren(db, Tag_cname, field, value, Cname, Tag_fieldname)
}

func TagSearch(db *mgo.Database, tag_slug string) ([]interface{}, error) {
	return patterns.FieldStartsWith(db, Tag_cname, "slug", tag_slug)
}
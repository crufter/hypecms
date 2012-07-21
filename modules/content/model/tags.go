package content_model

import (
	"github.com/opesun/hypecms/model/basic"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/slugify"
	"strings"
	"fmt"
)

const(
	Tag_fieldname = "_tags"					// goes to database
	Tag_fieldname_displayed = "tags"		// comes from user interface, it is in the rules
	Collection_name = "tags"
	Count_fieldname = "count"
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

func separateTags(db *mgo.Database, slug_sl []string) ([]bson.ObjectId, []string) {
	var i []interface{}
	db.C(Collection_name).Find(m{"slug":m{ "$in":slug_sl}}).Limit(0).All(&i)
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
	_, err := db.C(Collection_name).UpdateAll(m{"_id": m{"$in":ids}}, m{"$inc":m{Count_fieldname:1 }})
	return err
}

func dec(db *mgo.Database, ids []bson.ObjectId) error {
	_, err := db.C(Collection_name).UpdateAll(m{"_id": m{"$in":ids}}, m{"$inc":m{Count_fieldname:-1 }})
	return err
}

func insert(db *mgo.Database, slug_sl []string) []bson.ObjectId {
	ret := []bson.ObjectId{}
	for _, v := range slug_sl {
		id := bson.NewObjectId()
		tag := m{"_id": id, "slug":v, "name":v, Count_fieldname:0}
		db.C(Collection_name).Insert(tag)
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

func handleCount(db *mgo.Database, old_ids []bson.ObjectId, new_ids []bson.ObjectId) {
	dec_ids := diffIds(old_ids, new_ids)
	inc_ids := diffIds(new_ids, old_ids)
	inc(db, inc_ids)
	dec(db, dec_ids)
	fmt.Println(inc_ids, dec_ids)
}

func merge(a []bson.ObjectId, b []bson.ObjectId) []bson.ObjectId {
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

// $addToSet
// $pull

func handleTags(db *mgo.Database, dat map[string]interface{}, id string, mod string) {
	content := map[string]interface{}{}
	if mod != "insert" {
		content = find(db, basic.StripId(id))
	}
	tags_i, _ := dat[Tag_fieldname_displayed]
	delete(dat, Tag_fieldname_displayed)
	tags := tags_i.(string)					// Example: "Autók, Biciklik"
	tags_sl := strings.Split(tags, ",")
	slug_sl := []string{}
	for _, v := range tags_sl {
		slug := slugify.S(v)
		slug_sl = append(slug_sl, slug)
	}
	switch mod {
		case "insert":
			existing_ids, to_insert_slugs := separateTags(db, slug_sl)
			inserted_ids := insert(db, to_insert_slugs)
			old_ids := []bson.ObjectId{}
			new_ids := merge(existing_ids, inserted_ids)
			handleCount(db, old_ids, new_ids)
			dat[Tag_fieldname] = new_ids
		case "update":
			existing_ids, to_insert_slugs := separateTags(db, slug_sl)
			inserted_ids := insert(db, to_insert_slugs)
			old_ids := toIdSlice(content[Tag_fieldname].([]interface{}))
			new_ids := merge(existing_ids, inserted_ids)
			handleCount(db, old_ids, new_ids)
			dat[Tag_fieldname] = merge(old_ids, new_ids)	// TODO: handle duplication
		case "delete":
			old_ids := toIdSlice(content[Tag_fieldname].([]interface{}))
			new_ids := []bson.ObjectId{}
			handleCount(db, old_ids, new_ids)
		default:
			panic("Bad mode at handleTags.")
	}
}

func PullTag(db *mgo.Database, id string) error {
	
}

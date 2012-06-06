package content

import (
	//"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
)

type m map[string]interface{}

var Hooks = map[string]func(*context.Uni){
	"AD":        AD,
	"Front":     Front,
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
}

// Find slug value be given key.
func FindContent(db *mgo.Database, key, val string) (map[string]interface{}, bool) {
	query := make(bson.M)
	query[key] = val
	var v interface{}
	db.C("contents").Find(query).One(&v)
	if v == nil {
		return nil, false
	}
	return context.Convert(v).(map[string]interface{}), true
}

func Front(uni *context.Uni) {
	//uni.Put("article module runs")
	m, err := routep.Comp("/{slug}", uni.Req.URL.Path)
	if err == "" {
		content, found := FindContent(uni.Db, "slug", m["slug"])
		if found {
			uni.Put("found this shit")
			uni.Dat["_hijacked"] = true
			uni.Dat["_points"] = []string{"content"}
			uni.Dat["content"] = content
		}
	}

}

func Test(uni *context.Uni) {
	res := make(map[string]interface{})
	res["Front"] = jsonp.HasVal(uni.Opt, "Hooks.Front", "content")
	_, ok := jsonp.Get(uni.Opt, "Modules.Content")
	res["Modules"] = ok
	uni.Dat["_cont"] = res
}

func Install(uni *context.Uni) {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	content_options := m{
		"hello": 1,
	}
	uni.Db.C("options").Update(m{"_id": id}, m{"$addToSet": m{"Hooks.Front": "content"}, "$set": m{"Modules.content": content_options}})
}

func Uninstall(uni *context.Uni) {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	uni.Db.C("options").Update(m{"_id": id}, m{"$pull": m{"Hooks.Front": "content"}, "$unset": m{"Modules.content": 1}})
}

func AD(uni *context.Uni) {
	uni.Dat["_points"] = []string{"content/admin-index"}
}

func Back(uni *context.Uni) {

}

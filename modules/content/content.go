package content

import (
	//"nv/view"
	//"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
)

var Hooks = map[string]func(*context.Uni){
	"Test": Test,
}

// megadott kulcs alapján keresi a slugot
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

func HookFront(uni *context.Uni) {
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

func ArticleView() {

}

func EditView() {
}

func edit() {
}

func Edit() {
}

func Test(uni *context.Uni) {
	res := make(map[string]interface{})
	res["Front"] = jsonp.HasVal(uni.Opt, "Hooks.Front", "content")
	uni.Dat["_cont"] = res
}

func Install(uni *context.Uni) {
	//etries, ok := jsonp.Get(uni.Opt, "Modules.Display.Entries")
	//if !ok {
	//	uni.Put("there no entry points in display module")
	//	return
	//}
	//entries := map[string]interface{}{"file":"content"}
}

func init() {
	Hooks["Front"] = HookFront
}

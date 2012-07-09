package display_editor_model

import(
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/hypecms/model/basic"
	"labix.org/v2/mgo"
	"github.com/opesun/extract"
	"fmt"
)

type m map[string]interface{}

func New(db *mgo.Database, ev ifaces.Event, inp map[string][]string) error {
	name_sl, hn := inp["name"]
	if !hn {
		return fmt.Errorf("Can't save new display point: no name specified.")
	}
	name := name_sl[0]
	id := basic.CreateOptCopy(db)
	return db.C("options").Update(m{"_id":id}, m{"$set":m{ "Display-points." + name: m{}}})
}

func Save(db *mgo.Database, ev ifaces.Event, inp map[string][]string) error {
	rule := map[string]interface{}{"name":1, "prev_name":1, "queries":1}
	r := extract.New(rule)
	dat, err := r.Extract(inp)
	if err != nil {
		return err
	}
	if len(dat) != len(rule) {
		return fmt.Errorf("Missing fields:", basic.CalcMiss(rule, dat))
	}
	id := basic.CreateOptCopy(db)
	return db.C("options").Update(m{"_id":id}, m{"$set":m{ "Display-points." + dat["name"].(string): dat}})
}
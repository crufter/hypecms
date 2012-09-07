package display_editor_model

import (
	"fmt"
	"github.com/opesun/extract"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

type m map[string]interface{}

// Saves a new display point. A new display point simply an empty one, thus
// inp shall contain only one field: name.
func New(db *mgo.Database, ev ifaces.Event, inp map[string][]string) error {
	name_sl, hn := inp["name"]
	if !hn {
		return fmt.Errorf("Can't save new display point: no name specified.")
	}
	name := name_sl[0]
	id := basic.CreateOptCopy(db)
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"Display-points." + name: m{},
		},
	}
	return db.C("options").Update(q, upd)
}

// Updates an existing display point. We warn if an unkown key is sent too.
// See display_model.RunQueries for further explanation.
func Save(db *mgo.Database, ev ifaces.Event, inp map[string][]string) error {
	rule := map[string]interface{}{
		"name":      "must",
		"prev_name": "must",
		"queries":   "must",
	}
	r := extract.New(rule)
	dat, err := r.Extract(inp)
	if err != nil {
		return err
	}
	if len(dat) != len(rule) {
		return fmt.Errorf("Missing fields: %v", strings.Join(basic.CalcMiss(rule, dat), ", "))
	}
	que, err := jsonp.Decode(dat["queries"].(string))
	if err != nil {
		return err
	}
	que_m, ok := que.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Queries is not a map[string]interface{}.")
	}
	// TODO: this should be the job of extract module.
	for _, v := range que_m {
		for i, _ := range v.(map[string]interface{}) {
			switch i {
			case "c":
			case "q":
			case "sk":
			case "l":
			case "so":
			case "p":
			case "ex":
			default:
				return fmt.Errorf("Nonsensical field %v.", i)
			}
		}
	}
	upd := m{
		"$set": m{
			"Display-points." + dat["name"].(string) + ".queries": que_m,
		},
	}
	if dat["name"].(string) != dat["prev_name"].(string) {
		upd["$unset"] = m{
			"Display-points." + dat["prev_name"].(string) + ".queries": 1,
		}
	}
	id := basic.CreateOptCopy(db)
	q := m{"_id": id}
	return db.C("options").Update(q, upd)
}

func Delete(db *mgo.Database, ev ifaces.Event, inp map[string][]string) error {
	name_sl, hn := inp["name"]
	if !hn {
		return fmt.Errorf("Can't delete display point: no name specified.")
	}
	name := name_sl[0]
	id := basic.CreateOptCopy(db)
	q := m{"_id": id}
	upd := m{
		"$unset": m{
			"Display-points." + name: 1,
		},
	}
	return db.C("options").Update(q, upd)
}
func Install(db *mgo.Database, id bson.ObjectId) error {
	display_editor_options := m{}
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"Modules.display_editor": display_editor_options,
		},
	}
	return db.C("options").Update(q, upd)
}

func Uninstall(db *mgo.Database, id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$unset": m{
			"Modules.display_editor": 1,
		},
	}
	return db.C("options").Update(q, upd)
}

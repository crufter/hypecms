package content_model

import(
	"launchpad.net/mgo"
	//"launchpad.net/mgo/bson"
	ifaces "github.com/opesun/hypecms/interfaces"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/basic"
	"fmt"
)

func Insert(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string) error {
	id, hasid := dat["id"]
	if hasid && len(id[0]) > 0 {
		return fmt.Errorf("Can't insert an object wich already has an id.")
	}
	typ, hastype := dat["type"]
	if !hastype {
		return fmt.Errorf("No type when inserting content.")
	}
	ins_dat, extr_err := extract.New(rule).Extract(dat)
	if extr_err != nil {
		return extr_err
	}
	ins_dat["type"] = typ[0]
	return basic.Inud(db, ev, ins_dat, "contents", "insert", "")
}

func Update(db *mgo.Database, ev ifaces.Event, rule map[string]interface{}, dat map[string][]string) error {
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
	upd_dat["typ"] = typ[0]
	return basic.Inud(db, ev, upd_dat, "contents", "update", id[0])
}

func Delete(db *mgo.Database, ev ifaces.Event, id string) error {
	return basic.Inud(db, ev, nil, "contents", "delete", id)
}
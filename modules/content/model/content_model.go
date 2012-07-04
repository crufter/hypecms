package content_model

import(
	"launchpad.net/mgo"
	//"launchpad.net/mgo/bson"
	//"github.com/opesun/hypecms/model"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/api/scut"
	"fmt"
)

func Insert(db *mgo.Database, rule map[string]interface{}, dat map[string][]string) error {
	id, hasid := dat["id"]
	if hasid && len(id[0]) > 0 {
		return fmt.Errorf("Why send an ID at insert?")
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
	return scut.Inud(db, ins_dat, "contents", "insert", "")
}

func Update(db *mgo.Database, rule map[string]interface{}, dat map[string][]string) error {
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
	return scut.Inud(db, upd_dat, "contents", "update", id[0])
}

func Delete(db *mgo.Database, id string) error {
	return scut.Inud(db, nil, "contents", "delete", id)
}


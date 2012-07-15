package template_editor_model

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type m map[string]interface{}

func Install(db *mgo.Database, id bson.ObjectId) error {
	template_editor_options := m{
		// "example": "any value",
	}
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"Modules.template_editor": template_editor_options,
		},
	}
	return db.C("options").Update(q, upd)
}

func Uninstall(db *mgo.Database, id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$unset": m{
			"Modules.template_editor": 1,
		},
	}
	return db.C("options").Update(q, upd)
}
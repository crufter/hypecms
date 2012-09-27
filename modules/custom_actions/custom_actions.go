package custom_actions

import (
	"fmt"
	"github.com/opesun/hypecms/frame/context"
	ca_model "github.com/opesun/hypecms/modules/custom_actions/model"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo/bson"
)

type m map[string]interface{}

func (a *C) Execute() error {
	uni := a.uni
	action_name := uni.Req.Form["action"][0]
	action, has := jsonp.GetM(uni.Opt, "Modules.custom_actions.actions."+action_name)
	if !has {
		return fmt.Errorf("Can't find action %v in custom actions module.", action_name)
	}
	db := uni.Db
	user := uni.Dat["_user"].(map[string]interface{})
	opt := uni.Opt
	inp := map[string][]string(uni.Req.Form)
	typ := action["type"].(string)
	var r error
	switch typ {
	case "vote":
		r = ca_model.Vote(db, user, action, inp)
	case "respond_content":
		r = ca_model.RespondContent(db, user, action, inp, opt)
	default:
		r = fmt.Errorf("Unkown action %v at RunAction.", action_name)
	}
	return r
}


func (h *C) Install(id bson.ObjectId) error {
	custom_action_options := m{}
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"Modules.custom_actions": custom_action_options,
		},
	}
	return h.uni.Db.C("options").Update(q, upd)
}

func (h *C) Uninstall(id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$unset": m{
			"Modules.custom_actions": 1,
		},
	}
	return h.uni.Db.C("options").Update(q, upd)
}

type C struct {
	uni *context.Uni
}

func (c *C) Init(uni *context.Uni) {
	c.uni = uni
}
package custom_actions

import (
	"github.com/opesun/hypecms/api/context"
	ca_model "github.com/opesun/hypecms/modules/custom_actions/model"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo/bson"
	"fmt"
)

type m map[string]interface{}

var Hooks = map[string]interface{}{
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}

func Back(uni *context.Uni, action_name string) error {
	action, has := jsonp.GetM(uni.Opt, "Modules.custom_actions.actions." + action_name)
	if !has { return fmt.Errorf("Can't find action %v in custom actions module.", action_name) }
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

func Test(uni *context.Uni) error {
	return nil
}

func AD(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"custom_action/index"}
	// You can dispatch your different admin views here, based on url structure.
	return nil
}

func Install(uni *context.Uni, id bson.ObjectId) error {
	custom_action_options := m{
	}
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"Modules.custom_action": custom_action_options,
		},
	}
	return uni.Db.C("options").Update(q, upd)
}

func Uninstall(uni *context.Uni, id bson.ObjectId) error {
	q := m{"_id": id}
	upd := m{
		"$unset": m{
			"Modules.custom_action": 1,
		},
	}
	return uni.Db.C("options").Update(q, upd)
}
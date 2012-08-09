package custom_actions

import (
	"github.com/opesun/hypecms/api/context"
	ca_model "github.com/opesun/hypecms/modules/custom_actions/model"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo/bson"
	"fmt"
)

type m map[string]interface{}

var Hooks = map[string]func(*context.Uni) error {
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}

func Back(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
	act, has := jsonp.GetM(uni.Opt, "Modules.custom_actions.actions." + action)
	if !has { return fmt.Errorf("Can't find action %v in custom actions module.", action) }
	return ca_model.RunAction(uni.Db, uni.Dat["_user"].(map[string]interface{}), act, map[string][]string(uni.Req.Form), action)
}

func Test(uni *context.Uni) error {
	return nil
}

func AD(uni *context.Uni) error {
	uni.Dat["_points"] = []string{"custom_action/index"}
	// You can dispatch your different admin views here, based on url structure.
	return nil
}

func Install(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
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

func Uninstall(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	q := m{"_id": id}
	upd := m{
		"$unset": m{
			"Modules.custom_action": 1,
		},
	}
	return uni.Db.C("options").Update(q, upd)
}
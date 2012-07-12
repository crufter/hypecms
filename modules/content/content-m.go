package content

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/content/model"
	"github.com/opesun/jsonp"
	//"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"fmt"
)

type m map[string]interface{}

var Hooks = map[string]func(*context.Uni) error {
	"AD":        AD,
	"Front":     Front,
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
}

func Test(uni *context.Uni) error {
	front := jsonp.HasVal(uni.Opt, "Hooks.Front", "content")
	if !front {
		return fmt.Errorf("Not subscribed to front hook.")
	}
	return nil
}

func Install(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	content_options := m{
		"types": m {
			"blog": m{
				"rules" : m{
					"title": 1, "slug":1, "content": 1,
				},
			},
		},
	}
	return uni.Db.C("options").Update(m{"_id": id}, m{"$addToSet": m{"Hooks.Front": "content"}, "$set": m{"Modules.content": content_options}})
}

func Uninstall(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	return uni.Db.C("options").Update(m{"_id": id}, m{"$pull": m{"Hooks.Front": "content"}, "$unset": m{"Modules.content": 1}})
}

func SaveTypeConfig(uni *context.Uni) error {
	// id := scut.CreateOptCopy(uni.Db)
	return nil
}

func SaveConfig(uni *context.Uni) error {
	// id := scut.CreateOptCopy(uni.Db)
	return nil
}

// TODO: Move Ins, Upd, Del to other package since they can be used with all modules similar to content.
// TODO: Separate the shared processes of Insert/Update (type and rule checking, extracting)
func Ins(uni *context.Uni) error {
	typ, hastype := uni.Req.Form["type"]
	if !hastype {
		return fmt.Errorf("No type when inserting content.")
	}
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types." + typ[0] + ".rules")
	if !hasrule {
		return fmt.Errorf("Can't find content type rules " + typ[0])
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't insert content, you have no id.")
	}
	return content_model.Insert(uni.Db, uni.Ev, rule.(map[string]interface{}), map[string][]string(uni.Req.Form), uid.(bson.ObjectId))
}

// TODO: Separate the shared processes of Insert/Update (type and rule checking, extracting)
func Upd(uni *context.Uni) error {
	typ, hastype := uni.Req.Form["type"]
	if !hastype {
		return fmt.Errorf("No type when inserting content.")
	}
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types." + typ[0] + ".rules")
	if !hasrule {
		return fmt.Errorf("Can't find content type rules " + typ[0])
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't update content, you have no id.")
	}
	return content_model.Update(uni.Db, uni.Ev, rule.(map[string]interface{}), map[string][]string(uni.Req.Form), uid.(bson.ObjectId))
}

func Del(uni *context.Uni) error {
	id, has := uni.Req.Form["id"]
	if !has {
		return fmt.Errorf("No id sent from form when deleting content.")
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't update content, you have no id.")
	}
	return content_model.Delete(uni.Db, uni.Ev, id, uid.(bson.ObjectId))[0]	// HACK for now.
}

// Defaults to 100.
func AllowsComment(uni *context.Uni, inp map[string][]string, user_level int) (string, error) {
	typ, has_typ := inp["type"]
	if !has_typ {
		return "", fmt.Errorf("Can't find type when commenting.")
	}
	cont_opt, has := jsonp.GetM(uni.Opt, "Modules.content.types." + typ[0])
	if !has {
		return "", fmt.Errorf("Can't find options for content type " + typ[0])
	}
	err := content_model.AllowsComment(uni.Db, inp, cont_opt, uni.Dat["_user"].(map[string]interface{})["_id"].(bson.ObjectId), user_level)
	return typ[0], err
}

func InsertComment(uni *context.Uni) error {
	inp := map[string][]string(uni.Req.Form)
	typ, allow_err := AllowsComment(uni, inp, uLev(uni.Dat["_user"]))
	if allow_err != nil {
		return allow_err
	}
	comment_rule, hasrule := jsonp.GetM(uni.Opt, "Modules.content.types." + typ + ".comment_rules")
	if !hasrule {
		return fmt.Errorf("Can't find comment rules of content type " + typ)
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't insert comment, you have no id.")
	}
	return content_model.InsertComment(uni.Db, uni.Ev, comment_rule, inp, uid.(bson.ObjectId))
}

func UpdateComment(uni *context.Uni) error {
	inp := map[string][]string(uni.Req.Form)
	typ, allow_err := AllowsComment(uni, inp, uLev(uni.Dat["_user"]))
	if allow_err != nil {
		return allow_err
	}
	comment_rule, hasrule := jsonp.GetM(uni.Opt, "Modules.content.types." + typ + ".comment_rules")
	if !hasrule {
		return fmt.Errorf("Can't find comment rules of content type " + typ)
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't update comment, you have no id.")
	}
	return content_model.UpdateComment(uni.Db, uni.Ev, comment_rule, inp, uid.(bson.ObjectId))
}

func DeleteComment(uni *context.Uni) error {
	inp := map[string][]string(uni.Req.Form)
	_, allow_err := AllowsComment(uni, inp, uLev(uni.Dat["_user"]))
	if allow_err != nil {
		return allow_err
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't delete comment, you have no id.")
	}
	return content_model.DeleteComment(uni.Db, uni.Ev, inp, uid.(bson.ObjectId))
}

func uLev(useri interface{}) int {
	if useri == nil {
		return 0
	}
	user := useri.(map[string]interface{})
	ulev, has := user["level"]
	if !has {
		return 0
	}
	return int(ulev.(int))
}

func minLev(opt map[string]interface{}, op string) int {
	if v, ok := jsonp.Get(opt, "Modules.content." + op + "_level"); ok {
		return int(v.(float64))
	}
	return 300	// This is sparta.
}

func Back(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
	_, ok := jsonp.Get(uni.Opt, "Modules.content")
	if !ok {
		return fmt.Errorf("No content options.")
	}
	level := uni.Dat["_user"].(map[string]interface{})["level"].(int)
	if minLev(uni.Opt, action) > level {
		return fmt.Errorf("You have no rights to do content action " + action)
	}
	var r error
	switch action {
	case "insert":
		r = Ins(uni)
	case "update":
		r = Upd(uni)
	case "delete":
		r = Del(uni)
	case "insert_comment":
		r = InsertComment(uni)
	case "update_comment":
		r = UpdateComment(uni)
	case "delete_comment":
		r = DeleteComment(uni)
	case "save_config":
		r = SaveTypeConfig(uni)
	default:
		return fmt.Errorf("Can't find action named \"" + action + "\" in user module.")
	}
	return r 
}

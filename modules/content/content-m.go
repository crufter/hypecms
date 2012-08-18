package content

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/hypecms/modules/content/model"
	"github.com/opesun/jsonp"
	//"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"fmt"
	"strings"
)

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
	return content_model.Install(uni.Db, id)
}

func Uninstall(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	return content_model.Uninstall(uni.Db, id)
}

func SaveConfig(uni *context.Uni) error {
	// id := scut.CreateOptCopy(uni.Db)
	return nil
}

func prepareOp(uni *context.Uni, op string) (bson.ObjectId, string, error) {
	typ_s, hastype := uni.Req.Form["type"]
	if !hastype {
		return "", "", fmt.Errorf("No type when doing content op %v.", op)
	}
	typ := typ_s[0]
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return "", typ, fmt.Errorf("Can't %v content, you have no id.", op)
	}
	type_opt, _ := jsonp.GetM(uni.Opt, "Modules.content.types." + typ)
	allowed_err := content_model.AllowsContent(uni.Db, map[string][]string(uni.Req.Form), type_opt, uid.(bson.ObjectId), scut.ULev(uni.Dat["_user"]), op)
	if allowed_err != nil { return "", typ, allowed_err }
	return uid.(bson.ObjectId), typ, nil
}

// We never update drafts.
func SaveDraft(uni *context.Uni) error {
	post := uni.Req.Form
	typ_s, has_typ := post["type"]
	if !has_typ { return fmt.Errorf("No type when saving draft.") }
	typ := typ_s[0]
	content_type_opt, has_opt := jsonp.GetM(uni.Opt, "Modules.content.types." + typ)
	if !has_opt { return fmt.Errorf("Can't find options of content type %v.", typ) }
	allows := content_model.AllowsDraft(content_type_opt, scut.ULev(uni.Dat["_user"]), typ)
	if allows != nil { return allows }
	rules, has_rules := jsonp.GetM(uni.Opt, "Modules.content.types." + typ + ".rules")
	if !has_rules { return fmt.Errorf("Can't find rules of content type %v.", typ) }
	draft_id, err := content_model.SaveDraft(uni.Db, rules, map[string][]string(post))
	// Handle redirect.
	referer := uni.Req.Referer()
	is_admin := strings.Index(referer, "admin") != -1
	var redir string
	if err == nil {		// Go to the fresh draft if we succeeded to save it.
		redir = "/content/edit/" + typ + "_draft/" + draft_id.Hex()
	} else {			// Go back to the previous draft if we couldn't save the new one, or to the insert page if we tried to save a parentless draft.
		val, has := uni.Req.Form[content_model.Parent_draft_field]
		if has {
			redir = "/content/edit/" + typ + "_draft/" + val[0]
		} else {
			redir = "/content/edit/" + typ + "_draft/"
		}
	}
	if is_admin {
		redir = "/admin" + redir
	}
	uni.Dat["redirect"] = redir 
	return err
}

// TODO: Move Ins, Upd, Del to other package since they can be used with all modules similar to content.
func Insert(uni *context.Uni) error {
	uid, typ, prep_err := prepareOp(uni, "insert")
	if prep_err != nil { return prep_err }
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types." + typ + ".rules")
	if !hasrule {
		return fmt.Errorf("Can't find content type rules " + typ)
	}
	id, err := content_model.Insert(uni.Db, uni.Ev, rule.(map[string]interface{}), uni.Req.Form, uid)
	if err != nil { return err }
	// Handling redirect.
	is_admin := strings.Index(uni.Req.Referer(), "admin") != -1
	redir := "/content/edit/" + typ + "/" + id.Hex()
	if is_admin {
		redir = "/admin" + redir
	}
	uni.Dat["redirect"] = redir
	return nil
}

// TODO: Separate the shared processes of Insert/Update (type and rule checking, extracting)
func Update(uni *context.Uni) error {
	uid, typ, prep_err := prepareOp(uni, "insert")
	if prep_err != nil { return prep_err }
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types." + typ + ".rules")
	if !hasrule {
		return fmt.Errorf("Can't find content type rules " + typ)
	}
	err := content_model.Update(uni.Db, uni.Ev, rule.(map[string]interface{}), uni.Req.Form, uid)
	if err != nil { return err }
	// We must set redirect because it can come from draft edit too.
	is_admin := strings.Index(uni.Req.Referer(), "admin") != -1
	redir := "/content/edit/" + typ + "/" + basic.StripId(uni.Req.Form["id"][0])
	if is_admin {
		redir = "/admin" + redir
	}
	uni.Dat["redirect"] = redir
	return nil
}

func Delete(uni *context.Uni) error {
	uid, _, prep_err := prepareOp(uni, "insert")
	if prep_err != nil { return prep_err }
	id, has := uni.Req.Form["id"]
	if !has {
		return fmt.Errorf("No id sent from form when deleting content.")
	}
	return content_model.Delete(uni.Db, uni.Ev, id, uid)[0]	// HACK for now.
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
	inp := uni.Req.Form
	typ, allow_err := AllowsComment(uni, inp, scut.ULev(uni.Dat["_user"]))
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
	inp := uni.Req.Form
	typ, allow_err := AllowsComment(uni, inp, scut.ULev(uni.Dat["_user"]))
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
	_, allow_err := AllowsComment(uni, inp, scut.ULev(uni.Dat["_user"]))
	if allow_err != nil {
		return allow_err
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't delete comment, you have no id.")
	}
	return content_model.DeleteComment(uni.Db, uni.Ev, inp, uid.(bson.ObjectId))
}

func PullTags(uni *context.Uni) error {
	content_id := uni.Req.Form["content_id"][0]
	tag_id := uni.Req.Form["tag_id"][0]
	return content_model.PullTags(uni.Db, content_id, []string{tag_id})
	
}

func deleteTag(uni *context.Uni) error {
	tag_id := uni.Req.Form["tag_id"][0]
	return content_model.DeleteTag(uni.Db, tag_id)
}

func SaveTypeConfig(uni *context.Uni) error {
	// id := scut.CreateOptCopy(uni.Db)
	return content_model.SaveTypeConfig(uni.Db, map[string][]string(uni.Req.Form))
}

// TODO: Ugly name.
func SavePersonalTypeConfig(uni *context.Uni) error {
	user_id_i, has := jsonp.Get(uni.Dat, "_user._id")
	if !has { return fmt.Errorf("Can't find user id.") }
	user_id := user_id_i.(bson.ObjectId)
	return content_model.SavePersonalTypeConfig(uni.Db, map[string][]string(uni.Req.Form), user_id)
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
		if _, is_draft := uni.Req.Form["draft"]; is_draft {
			fmt.Println("///////////////////////////////////draft")
			r = SaveDraft(uni)
		} else {
			r = Insert(uni)
		}
	case "update":
		if _, is_draft := uni.Req.Form["draft"]; is_draft {
			fmt.Println("///////////////////////////////////draft")
			r = SaveDraft(uni)
		} else {
			r = Update(uni)
		}
	case "delete":
		r = Delete(uni)
	case "insert_comment":
		r = InsertComment(uni)
	case "update_comment":
		r = UpdateComment(uni)
	case "delete_comment":
		r = DeleteComment(uni)
	case "save_config":
		r = SaveTypeConfig(uni)
	case "pull_tags":
		r = PullTags(uni)
	case "delete_tag":
		r = deleteTag(uni)
	case "save_type_config":
		r = SaveTypeConfig(uni)
	case "save_personal_type_config":
		r = SavePersonalTypeConfig(uni)
	default:
		return fmt.Errorf("Can't find action named \"" + action + "\" in user module.")
	}
	return r 
}
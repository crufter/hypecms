package content

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/hypecms/model/patterns"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/content/model"
	"github.com/opesun/hypecms/modules/user"
	"github.com/opesun/jsonp"
	"fmt"
	"labix.org/v2/mgo/bson"
	"strings"
)

var Hooks = map[string]interface{}{
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

func Install(uni *context.Uni, id bson.ObjectId) error {
	return content_model.Install(uni.Db, id)
}

func Uninstall(uni *context.Uni, id bson.ObjectId) error {
	return content_model.Uninstall(uni.Db, id)
}

func SaveConfig(uni *context.Uni) error {
	// id := scut.CreateOptCopy(uni.Db)
	return nil
}

func allowsContent(uni *context.Uni, op string) (bson.ObjectId, string, error) {
	typ_s, hastype := uni.Req.Form["type"]
	if !hastype {
		return "", "", fmt.Errorf("No type when doing content op %v.", op)
	}
	typ := typ_s[0]
	if op != "insert" {
		if !content_model.Typed(uni.Db, patterns.ToIdWithCare(uni.Req.Form["id"][0]), typ) { 	// TODO: dont let it panic if not exists, return errror message.
			return "", "", fmt.Errorf("Content is not of type %v.", typ)
		}
	}
	auth_opts, ignore := user.AuthOpts(uni, "content.types." + typ, op)
	if ignore {
		return "", "", fmt.Errorf("Auth options should not be ignored.")
	}
	err, _ := user.AuthAction(uni, auth_opts)
	if err != nil {
		return "", "", err
	}
	uid_i, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return "", "", fmt.Errorf("Can't %v content, you have no id.", op)
	}
	uid := uid_i.(bson.ObjectId)
	user_level := scut.Ulev(uni.Dat["_user"])
	allowed_err := content_model.CanModifyContent(uni.Db, uni.Req.Form, 300, uid, user_level)
	if allowed_err != nil {
		return "", "", allowed_err
	}
	return uid, typ, nil
}

// We never update drafts.
func SaveDraft(uni *context.Uni) error {
	post := uni.Req.Form
	typ_s, has_typ := post["type"]
	if !has_typ {
		return fmt.Errorf("No type when saving draft.")
	}
	typ := typ_s[0]
	content_type_opt, has_opt := jsonp.GetM(uni.Opt, "Modules.content.types."+typ)
	if !has_opt {
		return fmt.Errorf("Can't find options of content type %v.", typ)
	}
	allows := content_model.AllowsDraft(content_type_opt, scut.Ulev(uni.Dat["_user"]), typ)
	if allows != nil {
		return allows
	}
	rules, has_rules := jsonp.GetM(uni.Opt, "Modules.content.types."+typ+".rules")
	if !has_rules {
		return fmt.Errorf("Can't find rules of content type %v.", typ)
	}
	draft_id, err := content_model.SaveDraft(uni.Db, rules, map[string][]string(post))
	// Handle redirect.
	referer := uni.Req.Referer()
	is_admin := strings.Index(referer, "admin") != -1
	var redir string
	if err == nil { // Go to the fresh draft if we succeeded to save it.
		redir = "/content/edit/" + typ + "_draft/" + draft_id.Hex()
	} else { // Go back to the previous draft if we couldn't save the new one, or to the insert page if we tried to save a parentless draft.
		draft_id, has_draft_id := uni.Req.Form[content_model.Parent_draft_field]
		if has_draft_id && len(draft_id[0]) > 0 {
			redir = "/content/edit/" + typ + "_draft/" + draft_id[0]
		} else if id, has_id := uni.Req.Form["id"]; has_id {
			redir = "/content/edit/" + typ + "/" + id[0]
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
	uid, typ, prep_err := allowsContent(uni, "insert")
	if prep_err != nil {
		return prep_err
	}
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types."+typ+".rules")
	if !hasrule {
		return fmt.Errorf("Can't find content type rules " + typ)
	}
	id, err := content_model.Insert(uni.Db, uni.Ev, rule.(map[string]interface{}), uni.Req.Form, uid)
	if err != nil {
		return err
	}
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
	uid, typ, prep_err := allowsContent(uni, "insert")
	if prep_err != nil {
		return prep_err
	}
	rule, hasrule := jsonp.Get(uni.Opt, "Modules.content.types."+typ+".rules")
	if !hasrule {
		return fmt.Errorf("Can't find content type rules " + typ)
	}
	err := content_model.Update(uni.Db, uni.Ev, rule.(map[string]interface{}), uni.Req.Form, uid)
	if err != nil {
		return err
	}
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
	uid, _, prep_err := allowsContent(uni, "insert")
	if prep_err != nil {
		return prep_err
	}
	id, has := uni.Req.Form["id"]
	if !has {
		return fmt.Errorf("No id sent from form when deleting content.")
	}
	return content_model.Delete(uni.Db, uni.Ev, id, uid)[0] // HACK for now.
}

func allowsComment(uni *context.Uni, op string) (string, error, error) {
	inp := uni.Req.Form
	user_level := scut.Ulev(uni.Dat["_user"])
	typ_s, has_typ := inp["type"]
	if !has_typ {
		return "", fmt.Errorf("Can't find content type when commenting."), nil
	}
	typ := typ_s[0]
	if op != "insert" {
		// We check this because they can lie about the type, sending a less strictly guarded type name and gaining access.
		if !content_model.Typed(uni.Db, bson.ObjectIdHex(inp["content_id"][0]), typ) {		// TODO: dont assume this exists.
			return "", fmt.Errorf("Content is not of type %v.", typ), nil
		}
	}
	auth_opts, ignore := user.AuthOpts(uni, "content.types." + typ, op + "_comment")
	if ignore {
		return "", fmt.Errorf("Auth options should not be ignored."), nil
	}
	err, puzzle_err := user.AuthAction(uni, auth_opts)
	if err != nil {
		return "", err, nil
	}
	var user_id bson.ObjectId
	user_id_i, has := jsonp.Get(uni.Dat, "_user._id")		// At this point the user will have a user id. TODO: except when the auth_opts is misconfigured.
	if !has {
		return "", fmt.Errorf("User has no id."), nil
	}
	if has {
		user_id = user_id_i.(bson.ObjectId)
	}
	if op != "insert" {
		err = content_model.CanModifyComment(uni.Db, inp, 300, user_id, user_level)		// TODO: remove hard-coded value.
	}
	return typ, err, puzzle_err
}

func InsertComment(uni *context.Uni) error {
	typ, allow_err, puzzle_err := allowsComment(uni, "insert")
	if allow_err != nil {
		return allow_err
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("You must have user id to comment.")
	}
	user_id := uid.(bson.ObjectId)
	comment_rule, hasrule := jsonp.GetM(uni.Opt, "Modules.content.types."+typ+".comment_rules")
	if !hasrule {
		return fmt.Errorf("Can't find comment rules of content type " + typ)
	}
	mf, has := jsonp.GetB(uni.Opt, "Modules.content.types."+typ+".moderate_comment")
	var moderate_first bool
	if (has && mf) || puzzle_err != nil {
		moderate_first = true
		uni.Dat["_cont"] = map[string]interface{}{"awaits-moderation": true}
	}
	inp := uni.Req.Form
	return content_model.InsertComment(uni.Db, uni.Ev, comment_rule, inp, user_id, moderate_first)
}

func UpdateComment(uni *context.Uni) error {
	typ, allow_err, _ := allowsComment(uni, "update")
	if allow_err != nil {
		return allow_err
	}
	comment_rule, hasrule := jsonp.GetM(uni.Opt, "Modules.content.types."+typ+".comment_rules")
	if !hasrule {
		return fmt.Errorf("Can't find comment rules of content type " + typ)
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't update comment, you have no id.")
	}
	inp := uni.Req.Form
	return content_model.UpdateComment(uni.Db, uni.Ev, comment_rule, inp, uid.(bson.ObjectId))
}

func DeleteComment(uni *context.Uni) error {
	_, allow_err, _ := allowsComment(uni, "delete")
	if allow_err != nil {
		return allow_err
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't delete comment, you have no id.")
	}
	return content_model.DeleteComment(uni.Db, uni.Ev, uni.Req.Form, uid.(bson.ObjectId))
}

func MoveToFinal(uni *context.Uni) error {
	return nil
}

func PullTags(uni *context.Uni) error {
	_, err, _ := allowsComment(uni, "update")
	if err != nil {
		return err
	}
	content_id := uni.Req.Form["id"][0]
	tag_id := uni.Req.Form["tag_id"][0]
	return content_model.PullTags(uni.Db, content_id, []string{tag_id})

}

func DeleteTag(uni *context.Uni) error {
	if scut.Ulev(uni.Dat["_user"]) < 300 {
		return fmt.Errorf("Only an admin can delete a tag.")
	}
	tag_id := uni.Req.Form["tag_id"][0]
	return content_model.DeleteTag(uni.Db, tag_id)
}

func SaveTypeConfig(uni *context.Uni) error {
	// id := scut.CreateOptCopy(uni.Db)
	return nil // Temp.
	return content_model.SaveTypeConfig(uni.Db, uni.Req.Form)
}

// TODO: Ugly name.
func SavePersonalTypeConfig(uni *context.Uni) error {
	return nil // Temp.
	user_id_i, has := jsonp.Get(uni.Dat, "_user._id")
	if !has {
		return fmt.Errorf("Can't find user id.")
	}
	user_id := user_id_i.(bson.ObjectId)
	return content_model.SavePersonalTypeConfig(uni.Db, uni.Req.Form, user_id)
}

func Back(uni *context.Uni, action string) error {
	var r error
	switch action {
	case "insert":
		if _, is_draft := uni.Req.Form["draft"]; is_draft {
			r = SaveDraft(uni)
		} else {
			r = Insert(uni)
		}
	case "update":
		if _, is_draft := uni.Req.Form["draft"]; is_draft {
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
		r = DeleteTag(uni)
	case "move_to_final":
		r = MoveToFinal(uni)
	case "save_type_config":
		r = SaveTypeConfig(uni)
	case "save_personal_type_config":
		r = SavePersonalTypeConfig(uni)
	default:
		return fmt.Errorf("Can't find action named \"" + action + "\" in user module.")
	}
	return r
}

package content

import (
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/hypecms/frame/misc/patterns"
	"github.com/opesun/hypecms/frame/misc/scut"
	"github.com/opesun/hypecms/modules/content/model"
	"github.com/opesun/hypecms/modules/user"
	"github.com/opesun/jsonp"
	"fmt"
	"labix.org/v2/mgo/bson"
)

const not_impl = "Not implemented yet."

func (h *H) Install(id bson.ObjectId) error {
	return content_model.Install(h.uni.Db, id)
}

func (h *H) Uninstall(id bson.ObjectId) error {
	return content_model.Uninstall(h.uni.Db, id)
}

// 
func (a *A) SaveConfig() error {
	// id := scut.CreateOptCopy(uni.Db)
	return fmt.Errorf(not_impl)
}

func (a *A) allowsContent(op string) (bson.ObjectId, string, error) {
	uni := a.uni
	var typ string
	if op == "insert" {
		typ = uni.Req.Form["type"][0]								// See TODO below.
	} else {
		content_id := patterns.ToIdWithCare(uni.Req.Form["id"][0])	// TODO: Don't let it panic if id does not exists, return descriptive error message.
		_typ, err := content_model.TypeOf(uni.Db, content_id)
		if err != nil {
			return "", "", err
		}
		typ = _typ
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

// We never update drafts, they are immutable.
func (a *A) saveDraft() error {
	uni := a.uni
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
	cont := map[string]interface{}{}
	if err == nil { 		// Go to the fresh draft if we succeeded to save it.
		cont["!type"] = typ+"_draft"
		cont["!id"] = draft_id.Hex()
	} else { // Go back to the previous draft if we couldn't save the new one, or to the insert page if we tried to save a parentless draft.
		draft_id, has_draft_id := uni.Req.Form[content_model.Parent_draft_field]
		if has_draft_id && len(draft_id[0]) > 0 {
			cont["!type"] = typ+"_draft"
			cont["!id"] = draft_id[0]
		} else if id, has_id := uni.Req.Form["id"]; has_id {
			cont["!type"] = typ
			cont["!id"] = id[0]
		} else {
			cont["!type"] = typ
			cont["!id"] = ""
		}
	}
	uni.Dat["_cont"] = cont
	return err
}

// Insert content.
// TODO: Move Ins, Upd, Del to other package since they can be used with all modules similar to content.
func (a *A) insert() error {
	uni := a.uni
	uid, typ, prep_err := a.allowsContent("insert")
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
	uni.Dat["_cont"] = map[string]interface{}{
		"!type": typ,
		"!id": id.Hex(),
	}
	return nil
}

func (a *A) Insert() error {
	if _, is_draft := a.uni.Req.Form["draft"]; is_draft {
		return a.saveDraft()
	}
	return a.insert()
}

// Update content.
// TODO: Consider separating the shared processes of Insert/Update (type and rule checking, extracting)
func (a *A) update() error {
	uni := a.uni
	uid, typ, prep_err := a.allowsContent("insert")
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
	draft_id, has_draft_id := uni.Req.Form[content_model.Parent_draft_field]
	content_id := uni.Req.Form["id"][0]
	if has_draft_id && len(draft_id[0]) > 0 {	// Coming from draft.
		// We must set redirect because it can come from draft edit too.
		uni.Dat["_cont"] = map[string]interface{}{
			"!type": 	typ,
			"!id":		content_id,
		}
	} else {
		uni.Dat["_cont"] = map[string]interface{}{
			"!type": typ,
		}
	}
	return nil
}

func (a *A) Update() error {
	if _, is_draft := a.uni.Req.Form["draft"]; is_draft {
		return a.saveDraft()
	}
	return a.update()
}

// Delete content.
func (a *A) Delete() error {
	uni := a.uni
	uid, _, prep_err := a.allowsContent("insert")
	if prep_err != nil {
		return prep_err
	}
	id, has := uni.Req.Form["id"]
	if !has {
		return fmt.Errorf("No id sent from form when deleting content.")
	}
	return content_model.Delete(uni.Db, uni.Ev, id, uid)[0] // HACK for now.
}

// Return values: content type, general (fatal) error, puzzle error
// Puzzle error is returned to support the decision of wether to put the comment into a moderation queue.
func (a *A) allowsComment(op string) (string, error, error) {
	uni := a.uni
	inp := uni.Req.Form
	user_level := scut.Ulev(uni.Dat["_user"])
	content_id := bson.ObjectIdHex(inp["content_id"][0])
	typ, err := content_model.TypeOf(uni.Db, content_id)
	if err != nil {
		return "", err, nil
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

func (a *A) InsertComment() error {
	uni := a.uni
	typ, allow_err, puzzle_err := a.allowsComment("insert")
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
	return content_model.InsertComment(uni.Db, uni.Ev, comment_rule, inp, user_id, typ, moderate_first)
}

func (a *A) UpdateComment() error {
	uni := a.uni
	typ, allow_err, _ := a.allowsComment("update")
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

func (a *A) DeleteComment() error {
	uni := a.uni
	_, allow_err, _ := a.allowsComment("delete")
	if allow_err != nil {
		return allow_err
	}
	uid, has_uid := jsonp.Get(uni.Dat, "_user._id")
	if !has_uid {
		return fmt.Errorf("Can't delete comment, you have no id.")
	}
	return content_model.DeleteComment(uni.Db, uni.Ev, uni.Req.Form, uid.(bson.ObjectId))
}

func (a *A) MoveToFinal() error {
	return nil
}

func (a *A) PullTags() error {
	uni := a.uni
	_, err, _ := a.allowsComment("update")
	if err != nil {
		return err
	}
	content_id := uni.Req.Form["id"][0]
	tag_id := uni.Req.Form["tag_id"][0]
	return content_model.PullTags(uni.Db, content_id, []string{tag_id})

}

func (a *A) DeleteTag() error {
	uni := a.uni
	if scut.Ulev(uni.Dat["_user"]) < 300 {
		return fmt.Errorf("Only an admin can delete a tag.")
	}
	tag_id := uni.Req.Form["tag_id"][0]
	return content_model.DeleteTag(uni.Db, tag_id)
}

func (a *A) SaveTypeConfig() error {
	uni := a.uni
	// id := scut.CreateOptCopy(uni.Db)
	return fmt.Errorf(not_impl)
	return content_model.SaveTypeConfig(uni.Db, uni.Req.Form)
}

// TODO: Ugly name.
func (a *A) SavePersonalTypeConfig() error {
	uni := a.uni
	return fmt.Errorf(not_impl) // Temp.
	user_id_i, has := jsonp.Get(uni.Dat, "_user._id")
	if !has {
		return fmt.Errorf("Can't find user id.")
	}
	user_id := user_id_i.(bson.ObjectId)
	return content_model.SavePersonalTypeConfig(uni.Db, uni.Req.Form, user_id)
}

type A struct {
	uni *context.Uni
}

func Actions(uni *context.Uni) *A {
	return &A{uni}
}

type H struct {
	uni *context.Uni
}

func Hooks(uni *context.Uni) *H {
	return &H{uni}
}
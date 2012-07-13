// Package skeleton implements a minimalistic but idiomatic plugin for HypeCMS.
package display_editor

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/display_editor/model"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"labix.org/v2/mgo/bson"
	"fmt"
	"strings"
	"encoding/json"
	"sort"
)

var Hooks = map[string]func(*context.Uni) error {
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}

type m map[string]interface{}

func New(uni *context.Uni) error {
	return display_editor_model.New(uni.Db, uni.Ev, map[string][]string(uni.Req.Form))
}

func Save(uni *context.Uni) error {
	return display_editor_model.Save(uni.Db, uni.Ev, map[string][]string(uni.Req.Form))
}

func Delete(uni *context.Uni) error {
	return display_editor_model.Delete(uni.Db, uni.Ev, map[string][]string(uni.Req.Form))
}

func Back(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
	var err error
	switch action {
	case "new":
		err = New(uni)
	case "save":
		err = Save(uni)
	case "delete":
		err = Delete(uni)
	default:
		return fmt.Errorf("Unkown display_editor action.")
	}
	return err
}

func Test(uni *context.Uni) error {
	return nil
}

func Search(uni *context.Uni) error {
	var search string
	if s, hass := uni.Req.Form["point-name"]; hass {
		search = s[0]
	}
	points, ok := jsonp.Get(uni.Opt, "Display-points")
	points_m := points.(map[string]interface{})
	has_points := false
	ps := []string{}
	if ok {
		for key, _ := range points_m {
			if search == "" || strings.Index(key, search) != -1 {
				ps = append(ps, key)
			}
			has_points = true
		}
	}
	uni.Dat["has_points"] = has_points
	sort.Strings(ps)
	uni.Dat["point_names"] = ps
	uni.Dat["search"] = search
	uni.Dat["_points"] = []string{"display_editor/search"}
	return nil
}

func Edit(uni *context.Uni, point_name string) error {
	point, ok := jsonp.GetM(uni.Opt, "Display-points." + point_name)
	if !ok {
		return fmt.Errorf("Can't find point named ", point_name)
	}
	var queries []interface{}
	if q, has := point["queries"]; has {
		queries = q.([]interface{})
	} else {
		queries = []interface{}{map[string]interface{}{}}
	}
	query_b, err := json.MarshalIndent(queries, "", "    ")
	var query string
	if err != nil {
		query = err.Error()
	} else {
		query = string(query_b)
	}
	uni.Dat["point"] = map[string]interface{}{"name": point_name, "queries": query}
	uni.Dat["_points"] = []string{"display_editor/edit"}
	return nil
}

func Help(uni *context.Uni, point_name string) error {
	uni.Dat["_points"] = []string{"display_editor/help"}
	return nil
}

func AD(uni *context.Uni) error {
	m, cerr := routep.Comp("/admin/display_editor/{view}/{param}", uni.P)
	if cerr != nil {
		return fmt.Errorf("Bad url at display editor AD.")
	}
	var err error
	switch m["view"] {
	case "":
		err = Search(uni)
	case "edit":
		err = Edit(uni, m["param"])
	case "help":
		err = Help(uni, m["param"])
	default:
		err = fmt.Errorf("Unkown view at display_editor admin: ", m["view"])
	}
	return err
}

func Install(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	display_editor_options := m{
	}
	q := m{"_id": id}
	upd := m{
		"$set": m{
			"Modules.display_editor": display_editor_options,
		},
	}
	return uni.Db.C("options").Update(q, upd)
}

func Uninstall(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	q := m{"_id": id}
	upd := m{
		"$unset": m{
			"Modules.display_editor": 1,
		},
	}
	return uni.Db.C("options").Update(q, upd)
}

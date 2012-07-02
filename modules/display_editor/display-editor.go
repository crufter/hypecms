// Package skeleton implements a minimalistic but idiomatic plugin for HypeCMS.
package display_editor

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/scut"
	"github.com/opesun/extract"
	"github.com/opesun/jsonp"
	"github.com/opesun/routep"
	"launchpad.net/mgo/bson"
	"fmt"
	"strings"
	"encoding/json"
)

type m map[string]interface{}

var Hooks = map[string]func(*context.Uni) error {
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}

func New(uni *context.Uni) error {
	name_sl, hn := uni.Req.Form["name"]
	if !hn {
		return fmt.Errorf("Can't save new display point: no name specified.")
	}
	name := name_sl[0]
	id := scut.CreateOptCopy(uni.Db)
	err := uni.Db.C("options").Update(m{"_id":id}, m{"$set":m{ "Display-points." + name: m{}}})
	return err
}

func Save(uni *context.Uni) error {
	rule := map[string]interface{}{"name":1, "queries":1}
	r := extract.New(rule)
	dat, err := r.ExtractForm(uni.Req.Form)
	if err != nil {
		return err
	}
	if len(dat) != len(rule) {
		return fmt.Errorf("Missing fields.")
	}
	id := scut.CreateOptCopy(uni.Db)
	err = uni.Db.C("options").Update(m{"_id":id}, m{"$set":m{ "Display-points." + dat["name"].(string): dat}})
	return err
}

func Back(uni *context.Uni) error {
	action := uni.Dat["_action"].(string)
	var err error
	switch action {
	case "new":
		err = New(uni)
	case "save":
		err = Save(uni)
	default:
		return fmt.Errorf("Unkown display_editor action.")
	}
	return err
}

func Test(uni *context.Uni) error {
	return nil
}

func Search(uni *context.Uni) {
	ps := []string{}
	var search string
	if s, hass := uni.Req.Form["point-name"]; hass {
		search = s[0]
	}
	points, ok := jsonp.Get(uni.Opt, "Display-points")
	points_m := points.(map[string]interface{})
	has_points := false
	if ok {
		for key, _ := range points_m {
			if search == "" || strings.Index(key, search) != -1 {
				ps = append(ps, key)
			}
			has_points = true
		}
	}
	uni.Dat["has_points"] = has_points
	uni.Dat["point_names"] = ps
	uni.Dat["search"] = search
	uni.Dat["_points"] = []string{"display_editor/search"}
}

func Edit(uni *context.Uni, point_name string) {
	point, ok := jsonp.Get(uni.Opt, "Display-points." + point_name)
	if !ok {
		uni.Put("Can't find point named " + point_name)
		return
	}
	query, _ := json.MarshalIndent(point, "", "    ")
	uni.Dat["point"] = map[string]interface{}{"name": point_name, "query": query}
	uni.Dat["_points"] = []string{"display_editor/edit"}
}

func Help(uni *context.Uni, point_name string) {
	uni.Dat["_points"] = []string{"display_editor/help"}
}

func AD(uni *context.Uni) error {
	m, err := routep.Comp("/admin/display_editor/{view}/{param}", uni.P)
	if err != nil {
		uni.Put("Bad url at display editor AD.")
		return nil
	}
	switch m["view"] {
	case "":
		Search(uni)
	case "edit":
		Edit(uni, m["param"])
	case "help":
		Help(uni, m["param"])
	}
	return nil
}

func Install(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	display_editor_options := m{
	}
	uni.Db.C("options").Update(m{"_id": id}, m{"$set": m{"Modules.display_editor": display_editor_options}})
	return nil
}

func Uninstall(uni *context.Uni) error {
	id := uni.Dat["_option_id"].(bson.ObjectId)
	uni.Db.C("options").Update(m{"_id": id}, m{"$unset": m{"Modules.display_editor": 1}})
	return nil
}

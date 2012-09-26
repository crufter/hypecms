package display_editor

import (
	"encoding/json"
	"fmt"
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/hypecms/modules/display_editor/model"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo/bson"
	"sort"
	"strings"
)

type m map[string]interface{}

func (a *A) New() error {
	return display_editor_model.New(a.uni.Db, a.uni.Ev, a.uni.Req.Form)
}

func (a *A) Save() error {
	return display_editor_model.Save(a.uni.Db, a.uni.Ev, a.uni.Req.Form)
}

func (a *A) Delete() error {
	return display_editor_model.Delete(a.uni.Db, a.uni.Ev, a.uni.Req.Form)
}

func (v *V) Index() error {
	uni := v.uni
	var search string
	if s, hass := uni.Req.Form["point-name"]; hass {
		search = s[0]
	}
	points, ok := jsonp.Get(uni.Opt, "Display-points")
	if !ok {
		// return fmt.Errorf("There is no \"Display-points\" field in the option document.") // Rather than freezing we easily recover here below.
		points = map[string]interface{}{}
	}
	// TODO: clean up here and make it more straightforward.
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
	uni.Dat["_points"] = []string{"display_editor/index"}
	return nil
}

func (v *V) Edit() error {
	uni := v.uni
	point_name := uni.Req.Form["point"][0]
	point, ok := jsonp.GetM(uni.Opt, "Display-points."+point_name)
	if !ok {
		return fmt.Errorf("Can't find point named ", point_name)
	}
	var queries map[string]interface{}
	if q, has := point["queries"]; has {
		queries = q.(map[string]interface{})
	} else {
		queries = map[string]interface{}{}
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

func (v *V) Help() error {
	v.uni.Dat["_points"] = []string{"display_editor/help"}
	return nil
}

func (h *H) Install(id bson.ObjectId) error {
	return display_editor_model.Install(h.uni.Db, id)
}

func (h *H) Uninstall(id bson.ObjectId) error {
	return display_editor_model.Uninstall(h.uni.Db, id)
}

type A struct {
	uni *context.Uni
}

func Actions(uni *context.Uni) *A {
	return &A{uni}
}

type V struct {
	uni *context.Uni
}

func Views(uni *context.Uni) *V {
	return &V{uni}
}

type H struct {
	uni *context.Uni
}

func Hooks(uni *context.Uni) *H {
	return &H{uni}
}

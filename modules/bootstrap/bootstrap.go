// Package bootstrap enables you (or others) to fork other HypeCMS istances from your instance.
// Used at hypecms.com
package bootstrap

import (
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	bm "github.com/opesun/hypecms/modules/bootstrap/model"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/routep"
	"strings"
	"github.com/opesun/numcon"
)

var Hooks = map[string]interface{}{
	// "Front":     	Front,
	"BeforeDisplay": 	BeforeDisplay,
	"Back":      		Back,
	"Install":   		Install,
	"Uninstall": 		Uninstall,
	"Test":      		Test,
	"AD":       		AD,
}

func BeforeDisplay(uni *context.Uni) {
	opt, has := jsonp.GetM(uni.Opt, "Modules.bootstrap")
	if !has {
		return
	}
	count, err := bm.SiteCount(uni.Db)
	if err != nil {
		return
	}
	max_cap := numcon.IntP(opt["max_cap"])
	ratio := float64(count)/float64(max_cap)
	perc := ratio * 100
	uni.Dat["capacity_percentage"] = perc
}

// Example bootstrap options: 			// All keys listed here are required to be able to ignite a site.
// {
//  "default_must": false,
//	"exec_abs": "c:/gowork/bin/hypecms",
//	"host_format": "%v.hypecms.com",
//	"max_cap": 500,
//	"proxy_abs": "c:/jsonfile.json",
//	"root_db": "hypecms",
//	"table_key": "proxy_table"
// }
func Ignite(uni *context.Uni) error {
	opt, has := jsonp.GetM(uni.Opt, "Modules.bootstrap")
	if !has {
		return fmt.Errorf("Bootstrap module is not installd properly.")
	}
	sitename, err := bm.Ignite(uni.Session, uni.Db, opt, uni.Req.Form)
	if err == nil {
		uni.Dat["_cont"] = map[string]interface{}{"sitename": sitename}
	}
	return err
}

// This function should be used only when neither of the processes are running, eg.
// when the server was restarted, or the forker process was killed, and all child processes died with it.
func StartAll(uni *context.Uni) error {
	opt, has := jsonp.GetM(uni.Opt, "Modules.bootstrap")
	if !has {
		return fmt.Errorf("Bootstrap module is not installd properly.")
	}
	if uni.Session == nil {
		return fmt.Errorf("This is not an admin instance.")
	}
	return bm.StartAll(uni.Db, opt)
}

func DeleteSite(uni *context.Uni) error {
	return bm.DeleteSite(uni.Db, uni.Req.Form)
}

func Back(uni *context.Uni, action string) error {
	switch action {
	case "ignite":
		return Ignite(uni)
	case "start-all":
		return StartAll(uni)
	default:
		return fmt.Errorf("Unkown action %v.")
	}
	return fmt.Errorf("Should not reach this point.")
}

func Test(uni *context.Uni) error {
	return fmt.Errorf("Not implemented yet.")
}

func filter(s []string, term string) []string {
	if len(term) == 0 {
		return s
	}
	ret := []string{}
	for _, v := range s {
		if strings.Index(v, term) != -1 {
			ret = append(ret, v)
		}
	}
	return ret
}

func Index(uni *context.Uni) error {
	not_admin := uni.Session == nil
	if not_admin {
		uni.Dat["not_admin"] = true
	}
	sitenames, err := bm.AllSitenames(uni.Db)
	if err != nil {
		return err
	}
	all := len(sitenames)
	var term string
	if len(uni.Req.Form["search"]) != 0 {
		term = uni.Req.Form["search"][0]
	}
	found := filter(sitenames, term)
	match := len(found)
	uni.Dat["all"] = all
	uni.Dat["match"] = match
	uni.Dat["sitenames"] = found
	uni.Dat["_points"] = []string{"bootstrap/index"}
	return nil
}

func AD(uni *context.Uni) error {
	ma, err := routep.Comp("/admin/bootstrap/{sub}", uni.P)
	if err != nil {
		return err
	}
	sub := ma["sub"]
	switch sub {
	case "":
		return Index(uni)
	default:
		return fmt.Errorf("Unkown view at bootstrap.")
	}
	return nil
}

func Install(uni *context.Uni, id bson.ObjectId) error {
	return bm.Install(uni.Session, uni.Db, id)
}

func Uninstall(uni *context.Uni, id bson.ObjectId) error {
	return bm.Uninstall(uni.Db, id)
}

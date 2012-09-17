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
)

var Hooks = map[string]interface{}{
	// "Front":     Front,
	"Back":      Back,
	"Install":   Install,
	"Uninstall": Uninstall,
	"Test":      Test,
	"AD":        AD,
}

// Example bootstrap options: 			// All keys listed here are required to be able to ignite a site.
// {
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
	return bm.Ignite(uni.Session, uni.Db, opt, uni.Req.Form)
}

// This function should be used only when neither of the processes are running, eg.
// when the server was restarted, or the forker process was killed.
func StartAll(uni *context.Uni) error {
	opt, has := jsonp.GetM(uni.Opt, "Modules.bootstrap")
	if !has {
		return fmt.Errorf("Bootstrap module is not installd properly.")
	}
	return bm.StartAll(uni.Db, opt)
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

func Index(uni *context.Uni) error {
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
	return bm.Install(uni.Db, id)
}

func Uninstall(uni *context.Uni, id bson.ObjectId) error {
	return bm.Uninstall(uni.Db, id)
}

package skeleton

import(
	"github.com/opesun/hypecms/frame/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/hypecms/frame/misc/scut"
	iface "github.com/opesun/hypecms/frame/interfaces"
	"github.com/opesun/hypecms/frame/composables/basics"
)

type C struct {
	basics.Basics
	uni *context.Uni
}

func (c *C) Init(uni *context.Uni) {
	c.uni = uni
}

func (c *C) New(a iface.Filter) error {
	scheme, ok := jsonp.GetM(c.uni.Opt, "nouns."+a.Subject()+".scheme")
	if !ok {
		scheme = map[string]interface{}{
			"info": 1,
			"name": 1,
		}
	}
	s, _ := scut.RulesToFields(scheme, nil)
	c.uni.Dat["scheme"] = s
	return nil
}

func (c *C) Doy(a iface.Filter) error {
	return nil
}
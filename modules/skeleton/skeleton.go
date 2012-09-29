package skeleton

import(
	"github.com/opesun/hypecms/frame/context"
	iface "github.com/opesun/hypecms/frame/interfaces"
	"github.com/opesun/hypecms/frame/composables/basics"
	"fmt"
)

type C struct {
	basics.Basics
}

func (c *C) Init(uni *context.Uni) {
	
}

func (c *C) Doy(a iface.Filter) error {
	return fmt.Errorf("loxxl")
}
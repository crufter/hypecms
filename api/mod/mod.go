// This package gets around the lack of dynamic code loading in Go.
package mod

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/modules/content"
	"github.com/opesun/hypecms/modules/skeleton"
	"github.com/opesun/hypecms/modules/user"
)

func GetHook(modname string, method string) func(*context.Uni) error {
	var r func(*context.Uni) error
	switch modname {
	case "content":
		r = content.Hooks[method]
	case "tag":

	case "user":
		r = user.Hooks[method]
	case "skeleton":
		r = skeleton.Hooks[method]
	}
	return r
}

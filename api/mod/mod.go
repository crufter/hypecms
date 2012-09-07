// This package gets around the lack of dynamic code loading in Go.
// There may be a better solution then putting every exported (in terms of the cms) function into a map.
package mod

import (
	"github.com/opesun/hypecms/modules/content"
	"github.com/opesun/hypecms/modules/custom_actions"
	"github.com/opesun/hypecms/modules/display_editor"
	"github.com/opesun/hypecms/modules/skeleton"
	"github.com/opesun/hypecms/modules/template_editor"
	"github.com/opesun/hypecms/modules/user"
)

var Modules = []string{
	"content",
	"user",
	"skeleton",
	"display_editor",
	"template_editor",
	"custom_actions",
}

func get(a interface{}, method string, map_only bool) interface{} {
	if map_only {
		return a
	}
	return a.(map[string]interface{})[method]
}

func getHook(modname string, method string, map_only bool) interface{} {
	var r interface{}
	switch modname {
	case "content":
		r = get(content.Hooks, method, map_only)
	case "user":
		r = get(user.Hooks, method, map_only)
	case "skeleton":
		r = get(skeleton.Hooks, method, map_only)
	case "display_editor":
		r = get(display_editor.Hooks, method, map_only)
	case "template_editor":
		r = get(template_editor.Hooks, method, map_only)
	case "custom_actions":
		r = get(custom_actions.Hooks, method, map_only)
	default:
		r = nil 	// panic("mod.Gethook cant find module " + modname)
	}
	return r
}

func GetHook(modname string, method string) interface{} {
	return getHook(modname, method, false)
}

// This is here for testing purposes only.
func GetHookMap(modname string) interface{} {
	return getHook(modname, "", true)
}

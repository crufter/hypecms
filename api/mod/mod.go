// This package gets around the lack of dynamic code loading in Go.
// There may be a better solution then putting every exported (in terms of the cms) function into a map.
package mod

// See
var Modules = map[string]map[string]interface{}{
}

func get(a interface{}, method string, map_only bool) interface{} {
	if map_only {
		return a
	}
	return a.(map[string]interface{})[method]
}

// Previously it was a big switch directly accessing the module Hooks,
// now its less efficient since it uses the Modules map, but this construct allows you to drop in files.
func getHook(modname string, method string, map_only bool) interface{} {
	modhooks, has := Modules[modname]
	if !has {
		return nil
	}
	return get(modhooks, method, map_only)
}

func GetHook(modname string, method string) interface{} {
	return getHook(modname, method, false)
}

// This is here for testing purposes only.
func GetHookMap(modname string) interface{} {
	return getHook(modname, "", true)
}

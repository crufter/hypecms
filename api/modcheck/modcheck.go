// This modules checks the exported hooks of the given modules and panics if they do not have the expected signature.
// This way we can avoid runtime problems.
package modcheck

import (
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/api/mod"
	"labix.org/v2/mgo/bson"
)

func adCheck(h interface{}) bool {
	_, ok := h.(func(*context.Uni) error)
	return ok
}

func installCheck(h interface{}) bool {
	_, ok := h.(func(*context.Uni, bson.ObjectId) error)
	return ok
}

func frontCheck(h interface{}) bool {
	_, ok := h.(func(*context.Uni, *bool) error)
	return ok
}

func backCheck(h interface{}) bool {
	_, ok := h.(func(*context.Uni, string) error)
	return ok
}

func testCheck(h interface{}) bool {
	_, ok := h.(func(*context.Uni) error)
	return ok
}

func add(a map[string][]string, b, c string) {
	_, has := a[b]
	if !has {
		a[b] = []string{}
	}
	a[b] = append(a[b], c)
}

func Check() {
	errs := map[string][]string{}
	for v, _ := range mod.Modules {		// Hehehe.
		m := mod.GetHookMap(v)
		_, ok := m.(map[string]interface{})
		if !ok {
			add(errs, v, "Hook map")
			continue
		}
		ad := mod.GetHook(v, "AD")
		if ad != nil && !adCheck(ad) {
			add(errs, v, "AD")
		}
		install := mod.GetHook(v, "Install")
		if install != nil && !installCheck(install) {
			add(errs, v, "Install")
		}
		uninstall := mod.GetHook(v, "Uninstall")
		if uninstall != nil && !installCheck(uninstall) {
			add(errs, v, "Uninstall")
		}
		front := mod.GetHook(v, "Front")
		if front != nil && !frontCheck(front) {
			add(errs, v, "Front")
		}
		back := mod.GetHook(v, "Back")
		if back != nil && !backCheck(back) {
			add(errs, v, "Back")
		}
		test := mod.GetHook(v, "Test")
		if test != nil && !testCheck(test) {
			add(errs, v, "Test")
		}
	}
	if len(errs) > 0 {
		fmt.Println("Next elements are mistyped:")
		fmt.Println("---")
		for i, v := range errs {
			fmt.Println(i)
			for _, x := range v {
				fmt.Println("	", x)
			}
			fmt.Println()
		}
		panic("Can't continue.")
	}
}

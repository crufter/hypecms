package mod

import "github.com/opesun/hypecms/modules/skeleton"

func init() {
	mods.register("skeleton", skeleton.C{})
}
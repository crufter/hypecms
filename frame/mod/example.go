package mod

import "github.com/opesun/hypecms/modules/example"

func init() {
	mods.register("example", example.C{})
}
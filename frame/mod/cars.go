package mod

import "github.com/opesun/hypecms/modules/cars"

func init() {
	mods.register("cars", cars.C{})
}
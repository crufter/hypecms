package mod

import ad "github.com/opesun/hypecms/modules/admin"

func init() {
	mods.register("admin", ad.C{})
}
package mod

import "github.com/opesun/hypecms/modules/user"

func init() {
	mods.register("user", user.C{})
}
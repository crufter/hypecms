package mod

import "github.com/opesun/hypecms/modules/users"

func init() {
	mods.register("users", users.C{})
}
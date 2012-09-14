package mod

import "github.com/opesun/hypecms/modules/user"

func init() {
	Modules["user"] = user.Hooks
}
package mod

import "github.com/opesun/hypecms/modules/user"

func init() {
	modules["user"] = dyn{Hooks: user.Hooks, Actions: user.Actions}
}
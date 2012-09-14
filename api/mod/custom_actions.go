package mod

import "github.com/opesun/hypecms/modules/custom_actions"

func init() {
	Modules["custom_actions"] = custom_actions.Hooks
}
package mod

import ca "github.com/opesun/hypecms/modules/custom_actions"

func init() {
	modules["custom_actions"] = dyn{Hooks: ca.Hooks, Actions: ca.Actions}
}
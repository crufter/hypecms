package mod

import c "github.com/opesun/hypecms/modules/content"

func init() {
	modules["content"] = dyn{Views: c.Views, Hooks: c.Hooks, Actions: c.Actions}
}
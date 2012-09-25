package mod

import te "github.com/opesun/hypecms/modules/template_editor"

func init() {
	modules["template_editor"] = dyn{Views: te.Views, Hooks: te.Hooks, Actions: te.Actions}
}
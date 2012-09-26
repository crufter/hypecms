package mod

import de "github.com/opesun/hypecms/modules/display_editor"

func init() {
	modules["display_editor"] = dyn{Views: de.Views, Hooks: de.Hooks, Actions: de.Actions}
}
package mod

import "github.com/opesun/hypecms/modules/display_editor"

func init() {
	Modules["display_editor"] = display_editor.Hooks
}
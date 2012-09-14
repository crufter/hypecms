package mod

import "github.com/opesun/hypecms/modules/template_editor"

func init() {
	Modules["template_editor"] = template_editor.Hooks
}
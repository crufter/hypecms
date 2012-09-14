package mod

import "github.com/opesun/hypecms/modules/content"

func init() {
	Modules["content"] = content.Hooks
}
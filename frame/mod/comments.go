package mod

import "github.com/opesun/hypecms/modules/comments"

func init() {
	mods.register("comments", comments.C{})
}
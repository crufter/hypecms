package mod

import "github.com/opesun/hypecms/modules/skeleton"

func init() {
	Modules["skeleton"] = skeleton.Hooks
}
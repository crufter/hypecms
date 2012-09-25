package mod

import "github.com/opesun/hypecms/modules/skeleton"

func init() {
	modules["skeleton"] = dyn{Hooks: skeleton.Hooks}
}
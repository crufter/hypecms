package mod

import "github.com/opesun/hypecms/modules/skeleton"

func init() {
	modules["skeleton"] = dyn{Views: skeleton.Views, Hooks: skeleton.Hooks}
}
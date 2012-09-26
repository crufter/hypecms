package mod

import "github.com/opesun/hypecms/modules/example"

func init() {
	modules["example"] = dyn{Hooks: example.Hooks}
}
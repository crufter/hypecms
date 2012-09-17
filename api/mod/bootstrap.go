package mod

import "github.com/opesun/hypecms/modules/bootstrap"

func init() {
	Modules["bootstrap"] = bootstrap.Hooks
}
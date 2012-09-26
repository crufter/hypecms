package mod

import ad "github.com/opesun/hypecms/modules/admin"

func init() {
	modules["admin"] = dyn{Actions: ad.Actions, Views:ad.Views}
}
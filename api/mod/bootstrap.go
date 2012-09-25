package mod

// bs, hahahahaha.
import bs "github.com/opesun/hypecms/modules/bootstrap"

func init() {
	modules["bootstrap"] = dyn{Hooks: bs.Hooks, Actions: bs.Actions}
}
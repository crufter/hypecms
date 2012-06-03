// A contextus gyakorlatilag a rendszer lelke, mindent el tudunk ez alól érni, ám csak a kontrollereknak adjuk tovább,
// tartózkodjunk attól, hogy a modell rétegnek továbbpasszoljuk.
package context

import (
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"net/http"
)

// Univerzum, csak és kizárólag kontrollereknek szabad elérniük.
// Tartózkodjunk attól, hogy a modell rétegik lepasszoljuk, esetleg csak a DB-t, ahol szükséges.
type Uni struct {
	Db    *mgo.Database
	W     http.ResponseWriter
	Req   *http.Request
	Paths []string // Path slice, értéke Req.URL.Path "/"-enként szétszplitelt
	P     string
	Opt   map[string]interface{} // Az adatbázisban tárolt beállításokat itt érjük el. (Minden modul minden beállítása itt van egyben)
	Dat   map[string]interface{} // Mivel a rendszer moduláris, kell egy kommunikációs csatorna, amiből minden modul ír/olvas. A Dat erre van fenntartva.
	Put   func(...interface{})   // kényelmi funkció fejlesztéshez http kimenet egyszerűsítve, gyorsan...
	Root  string                 // Abszolút elérési útvonala a project mappának.
}

// Rekurzívan, többszörösen egymásba ágyazott bson.M-eket átalakít sima map[string]interface{}-re.
// Rog Peppe írta
func Convert(x interface{}) interface{} {
	if x, ok := x.(bson.M); ok {
		for key, val := range x {
			x[key] = Convert(val)
		}
		return (map[string]interface{})(x)
	}
	return x
}

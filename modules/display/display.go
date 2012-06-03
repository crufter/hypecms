// A display egy megjelenítési pontot ("entry") próbál lefuttatni. Ez gyakorlatilag a rendszer view/nézet része.
// Bármi ami megjelenik a képernyőn egy átlagos felhasználó számára (leszámítva a hibaüzeneteket, vagy a kontrollerek eredményét, ha nem redirectelünk, hane json=true-val kiíratjuk azt)
// ezen keresztül jelenítődik meg.
package display

import (
	"fmt"
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/jsonp"
	"github.com/opesun/require"
	"html/template"
	"path/filepath"
	"strings"
)

// Lefut, ha hiba történik a template file értelmezése/megjelenítése közben.
func displErr(uni *context.Uni) {
	r := recover()
	if r != nil {
		uni.Put("The template file is a pile of shit and contains malformed data what the html/template pkg can't handle.")
	}
}

// "user/search" 	oldal templatejében keres először, ha nincs felülírva akkor előszedi az alapot user modul tpl mappájából
// "content"		csak az oldal templatejében
// Ez amúgy nem olyan jó: pl "rolunk/lali" is úgy néz ki, mint egy adminos gecmájer.

// Megjeleníti az adott template filet. Ha a template file filep-ja "rolunk/lali" akkor is működik, de a require-él vegyük figyelembe azt, hogy
// a require header.t nem a "rolunk/header.t"-t fogja betölteni, hanem a sima "header.t"-t.
func DisplayTemplate(uni *context.Uni, filep string) string {
	tpl, has_tpl := uni.Opt["Template"]
	if !has_tpl {
		tpl = "default"
	}
	templ := tpl.(string)
	_, priv := uni.Opt["TplIsPrivate"]
	var ttype string
	if priv {
		ttype = "private"
	} else {
		ttype = "public"
	}
	file, err := require.RSimple(filepath.Join(uni.Root, "templates", ttype, templ), filep+".tpl")
	if err == "" {
		uni.Dat["_tpl"] = "/templates/" + ttype + "/" + templ + "/"
		t, _ := template.New("template_name").Parse(string(file))
		_ = t.Execute(uni.W, uni.Dat)
		return ""
	}
	return "cant find template file " + `"` + filep + `"`
}

// A fallback function megpróbálja előbányászni a nem található template filet a modulok/pl mappából. Ez csak admin fájloknál használatos efektíven,
// Mert így az admin felületet is felül tudja írni egy adott template. "Rendes" template fájloknál kevésbé hasznos egy callback, mivel nincs header-je, vagy különbözik
// a template által használatosnál.
func DisplayFallback(uni *context.Uni, filep string) string {
	if strings.Index(filep, "/") != -1 {
		p := strings.Split(filep, "/")
		if len(p) >= 2 {
			// csúnya duplikació
			file, err := require.RSimple(filepath.Join(uni.Root, "modules", p[0], "tpl"), strings.Join(p[1:], "/")+".tpl")
			if err == "" {
				uni.Dat["_tpl"] = "/modules/" + p[0] + "/tpl/"
				t, _ := template.New("template_name").Parse(string(file))
				_ = t.Execute(uni.W, uni.Dat)
				return ""
			}
			return "cant find fallback template file " + `"` + filep + `"`
		}
		return "fallback filep is too long"
	}
	return "fallback filep contains no slash, so there nothing to fall back"
}

// Beolvassa a filet és kiírja képernyőre, de előtte a lefuttatja rajta a "require" modult, és a Go template modulját is.
func DisplayFile(uni *context.Uni, filep string) {
	defer displErr(uni)
	err := DisplayTemplate(uni, filep)
	if err != "" {
		err_f := DisplayFallback(uni, filep)
		if err_f != "" {
			uni.Put(err, "\n", err_f)
		}
	}
}

// Ha hiba történik a queryk futtatása közben...
func queryErr(uni *context.Uni) {
	r := recover()
	fmt.Println(r)
	uni.Put("shit happened while running queries")
}

// A runQueries a megjelenítési ponthoz ("entry"), az adatbázisban letárolt lekérdezéseket lefuttatja.
// Ez egyelőre akkora runtime panicet dob ha nem megfelelő az adat az adatbázisban hogy leszakad a fejed, ezért is van defer... todo: javítani
func runQueries(uni *context.Uni, queries []map[string]interface{}) {
	defer queryErr(uni)
	qs := make(map[string]interface{})
	for _, v := range queries {
		q := uni.Db.C(v["c"].(string)).Find(v["q"])
		if skip, skok := v["sk"]; skok {
			q.Skip(skip.(int))
		}
		if limit, lok := v["l"]; lok {
			q.Limit(limit.(int))
		}
		if sort, sook := v["so"]; sook {
			q.Sort(jsonp.ToStringSlice(sort))
		}
		var res interface{}
		q.All(&res)
		qs[v["n"].(string)] = res
	}
	uni.Dat["queries"] = qs
}

// Itt keződik a package futása, megpróbál elővenni egy megjelenítési pontot, ha nem találja, akkor viszont csak simán a megjelenítési pont nevével megegyező .tpl filet fog képernyőre íratni.
// Kivétel: "/" file, mert az "/index"-re fog módosulni
// Változni fog, nagyon.
func D(uni *context.Uni) {
	points, points_exist := uni.Dat["_points"]
	var point, filep string
	if points_exist {
		point = points.([]string)[0]
		queries, queries_exists := jsonp.Get(uni.Opt, "Modules.Display.Points."+point+".queries")
		if queries_exists {
			qslice, ok := queries.([]map[string]interface{})
			if ok {
				runQueries(uni, qslice)
			}
		}
		filep = point
		// Ha nincs point
	} else {
		p := uni.Req.URL.Path
		if p == "/" {
			filep = "index"
		} else {
			filep = p
		}
	}
	DisplayFile(uni, filep)
}

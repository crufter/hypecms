package display_model

// All functions which can be called from templates resides here.

import(
	"strings"
	"github.com/opesun/jsonp"
)

func get(dat map[string]interface{}, s ...string) interface{} {
	if len(s) > 0 {
		if len(s[0]) > 0 {
			if string(s[0][0]) == "$" {
				s[0] = s[0][1:]
			}
		}
	}
	access := strings.Join(s, ".")
	val, has := jsonp.Get(dat, access)
	if !has { return access }
	return val
}

// date is inte
func date(timestamp int, format string) string {
	return ""
}

// We must recreate this map each time because map access is not threadsafe.
func Builtins(dat map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{
		"get": func(s ...string) interface{} {
			return get(dat, s...)
		},
		"date": date,
	}
	return ret
}
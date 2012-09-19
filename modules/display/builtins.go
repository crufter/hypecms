package display

// All functions which can be called from templates resides here.

import (
	"github.com/opesun/hypecms/api/context"
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/hypecms/modules/user"
	"github.com/opesun/jsonp"
	"html/template"
	"reflect"
	"strings"
	"time"
	"strconv"
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
	if !has {
		return access
	}
	return val
}

func date(timestamp int64, format ...string) string {
	var form string
	if len(format) == 0 {
		form = "2006.01.02 15:04:05"
	} else {
		form = format[0]
	}
	t := time.Unix(timestamp, 0)
	return t.Format(form)
}

func isMap(a interface{}) bool {
	v := reflect.ValueOf(a)
	switch kind := v.Kind(); kind {
	case reflect.Map:
		return true
	}
	return false
}

func eq(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func showPuzzles(uni *context.Uni, mod_name, action_name string) string {
	str, err := user.ShowPuzzlesPath(uni, mod_name, action_name)
	if err != nil {
		return err.Error()
	}
	return str
}

func html(s string) template.HTML {
	return template.HTML(s)
}

func formatFloat(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}

// We must recreate this map each time because map write is not threadsafe.
// Write will happen when a hook modifies the map (hook call is not implemented yet).
func builtins(uni *context.Uni) map[string]interface{} {
	dat := uni.Dat
	user := uni.Dat["_user"]
	ret := map[string]interface{}{
		"get": func(s ...string) interface{} {
			return get(dat, s...)
		},
		"date": date,
		"solved_puzzles": func() bool {
			return scut.SolvedPuzzles(user)
		},
		"is_stranger": func() bool {
			return scut.IsStranger(user)
		},
		"is_guest": func() bool {
			return scut.IsGuest(user)
		},
		"is_registered": func() bool {
			return scut.IsRegistered(user)
		},
		"is_moderator": func() bool {
			return scut.IsModerator(user)
		},
		"is_admin": func() bool {
			return scut.IsAdmin(user)
		},
		"is_map": isMap,
		"eq": eq,
		"show_puzzles": func(a, b string) string {
			return showPuzzles(uni, a, b)
		},
		"html": html,
		"format_float": formatFloat,
	}
	return ret
}

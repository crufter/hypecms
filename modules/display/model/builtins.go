package display_model

// All functions which can be called from templates resides here.

import (
	"github.com/opesun/hypecms/model/scut"
	"github.com/opesun/jsonp"
	"reflect"
	"strings"
	"time"
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

func date(timestamp int64, format string) string {
	t := time.Unix(timestamp, 0)
	return t.Format(format)
}

func eq(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// We must recreate this map each time because map access is not threadsafe.
func Builtins(dat map[string]interface{}) map[string]interface{} {
	user := dat["_user"]
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
		"eq": eq,
	}
	return ret
}

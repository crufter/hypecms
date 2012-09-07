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

func isGuest(user interface{}) bool {
	return scut.IsGuest(user)
}

func isRegistered(user interface{}) bool {
	return scut.IsRegistered(user)
}

func isModerator(user interface{}) bool {
	return scut.IsModerator(user)
}

func isAdmin(user interface{}) bool {
	return scut.IsAdmin(user)
}

func eq(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// We must recreate this map each time because map access is not threadsafe.
func Builtins(dat map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{
		"get": func(s ...string) interface{} {
			return get(dat, s...)
		},
		"date": date,
		"is_guest": func() bool {
			return isGuest(dat["_user"])
		},
		"is_registered": func() bool {
			return isRegistered(dat["_user"])
		},
		"is_moderator": func() bool {
			return isModerator(dat["_user"])
		},
		"is_admin": func() bool {
			return isAdmin(dat["_user"])
		},
		"eq": eq,
	}
	return ret
}

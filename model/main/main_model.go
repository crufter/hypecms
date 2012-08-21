package main_model

import(
	"labix.org/v2/mgo"
	"sync"
	"encoding/json"
	"fmt"
)

type m map[string]interface{}

const(
	cached_opt_inv				= "The cached options string is not a valid JSON." 									// TODO: Maybe we should try to recover from here.
	cant_unmarshal				= "Can't unmarshal freshly encoded option document."
	cant_encode_config			= "Can't encode config. - No way this should happen anyway."
)

func set(c map[string]string, key, val string) {
	mut := new(sync.Mutex)
	mut.Lock()
	c[key] = val
	mut.Unlock()
}

// mutex locked map get
func has(c map[string]string, str string) (string, bool) {
	mut := new(sync.Mutex)
	mut.Lock()
	v, ok := c[str]
	mut.Unlock()
	return v, ok
}

// This is an artifact from the old version, we don't need the map here.
// Leave it as is, we may migrate back to the "multiple http servers from one process" approach. *1
var cache = make(map[string]string)

func HandleConfig(db *mgo.Database, host string, cache_it bool) (map[string]interface{}, string, error) {
	host = "anything"	// See *1
	ret := map[string]interface{}{}
	var ret_str string
	if val, ok := has(cache, host); cache_it && ok {
		ret_str = val
		var v interface{}
		json.Unmarshal([]byte(val), &v)
		if v == nil {
			return nil, "", fmt.Errorf(cached_opt_inv)
		}
		ret = v.(map[string]interface{})
		delete(ret, "_id")
	} else {
		var res interface{}
		db.C("options").Find(nil).Sort("-created").Limit(1).One(&res)
		if res == nil {
			res = m{}
			db.C("options").Insert(res)
		}
		enc, merr := json.Marshal(res)
		if merr != nil {
			return nil, "", fmt.Errorf(cant_encode_config)
		}
		str := string(enc)
		ret_str = str
		set(cache, host, str)
		var v interface{}
		json.Unmarshal([]byte(str), &v)
		if v == nil {
			return nil, "", fmt.Errorf(cant_unmarshal)
		}
		ret = v.(map[string]interface{})
		delete(ret, "_id")
	}
	return ret, ret_str, nil
}
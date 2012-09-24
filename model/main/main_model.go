// This package contains helper functions which run at the top level of the system - at main.go
package main_model

import (
	"encoding/json"
	"fmt"
	"labix.org/v2/mgo"
	"sync"
	"time"
)

type m map[string]interface{}

const (
	cached_opt_inv     = "The cached options string is not a valid JSON." // TODO: Maybe we should try to recover from here.
	cant_unmarshal     = "Can't unmarshal freshly encoded option document."
	cant_encode_config = "Can't encode config. - No way this should happen anyway."
)

// mutex locked map set
// Puts the JSON encoded string version of the option document to cache.
func set(c map[string]string, key, val string) {
	mut := new(sync.Mutex)
	mut.Lock()
	defer mut.Unlock()
	c[key] = val
}

// mutex locked map get
// Gets the JSON encoded string version of the option document from the cache.
func has(c map[string]string, str string) (string, bool) {
	mut := new(sync.Mutex)
	mut.Lock()
	defer mut.Unlock()
	v, ok := c[str]
	return v, ok
}

// This is an artifact from the old version, we don't need the map here.
// Leave it as is, we may migrate back to the "multiple http servers from one process" approach. *1
var cache = make(map[string]string)

// Loads the freshest option document from the database, and caches it if cache_it is true.
// Returns both the map[string]interface{} which comes from the database directly, and a JSON encoded string version too.
// The string version is being returned to be able to serve a version of the option document which is 100% untampered.
// (Assuming that the string is stored as private and only a copy of it can be retrieved)
//
// The data is also stored in the cache as a string to provide its immutability.
// (One pageload this way can't mess up the option document for the next.)
func HandleConfig(db *mgo.Database, host string, cache_it bool) (map[string]interface{}, string, error) {
	host = "anything" // See *1
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
		var res []interface{}
		err := db.C("options").Find(nil).Sort("-created").Limit(1).All(&res)
		if err != nil {
			return nil, "", err
		}
		var fresh_opt interface{}
		if len(res) == 0 {
			fresh_opt = m{}
			db.C("options").Insert(m{"created":time.Now().UnixNano()})		// Intentionally skipping error here.
		} else {
			fresh_opt = res[0]
		}
		enc, merr := json.Marshal(fresh_opt)
		if merr != nil {
			return nil, "", fmt.Errorf(cant_encode_config)
		}
		str := string(enc)
		ret_str = str
		if cache_it {
			set(cache, host, str)
		}
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

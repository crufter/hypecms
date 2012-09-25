// Context contains the type Uni. An instance of this type is passed to the modules when routing the control to them.
package context

import (
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/hypecms/interfaces"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo"
	"net/http"
	"strings"
	"reflect"
	"fmt"
)

// General context for the application.
type Uni struct {
	Session 	*mgo.Session
	Db      	*mgo.Database
	W       	http.ResponseWriter
	Req     	*http.Request
	secret  	string                 		// Used for things like encryption/decryption. Basically a permanent random data.
	P       	string                 		// Path string
	Paths   	[]string               		// Path slice, contains the url (after the domain) splitted by "/"
	opt     	string                 		// Original string representation of the option, if one needs a version which is guaranteedly untinkered.
	Opt     	map[string]interface{} 		// Freshest options from database.
	Dat     	map[string]interface{} 		// General communication channel.
	Put     	func(...interface{})   		// Just a convenience function to allow fast output to http response.
	Root    	string                 		// Absolute path of the application.
	Ev      	*Ev
	Caller 		interfaces.Caller
}

// Set only once.
func (u *Uni) SetOriginalOpt(s string) {
	if u.opt == "" {
		u.opt = s
	}
}

func (u *Uni) OriginalOpt() string {
	return u.opt
}

// Maybe we should not even return the secret, because a badly written module can make publish it.
// Or, we could serve different values to different packages.
// That makes the encrypted values noncompatible across packages though.
func (u *Uni) Secret() string {
	return u.secret
}

// Set only once.
func (u *Uni) SetSecret(s string) {
	if u.secret == "" {
		u.secret = s
	}
}

// With the help of this type it's possible for the model to not have direct access to everything (*context.Uni), but still trigger events,
// which in turn will result in hooks (which may need access to everything) being called.
// Ev is an implementation of the interface Event fround in "github.com/opesun/hypecms/interfaces"
type Ev struct {
	uni    *Uni
}

// Return all hooks modules subscribed to a path.
func all(e *Ev, path string) []string {
	modnames, ok := jsonp.GetS(e.uni.Opt, "Hooks." + path)
	if !ok {
		return nil
	}
	ret := []string{}
	for _, v := range modnames {
		ret = append(ret, v.(string))
	}
	return ret
}

// Trigger calls hooks subscribed to eventname, passes *Uni as a first parameter if the given hook needs it (eg *context.Uni
// is defined as its first parameter), and params... if they are given.
//
// Example eventname: "content.insert"
// Note that, different subscriptions should not be created for subsets of functionality,
// eg: "content.blog.insert" (where blog is a content type) should not be used, because we build the hook function name from the access path, eg:
// content.insert => ContentInsert may be a static, valid hookname, but ContentBlogInsert not (a module author can't know that the type will be blog or
// cats ahead of time).
//
// Filtering can be done inside ContentInsert if one module wants to act only on certain information (in the example case on certain content types).
func (e *Ev) Trigger(eventname string, params ...interface{}) {
	e.trigger(eventname, nil, params...)
}

// Calls all hooks subscribed to eventname, with params, feeding the output of every hook into stopfunc.
// Stopfunc's argument signature must match the signatures of return values of the called hooks.
// Stopfunc must return a boolean value. A boolean value of true stops the iteration.
// Iterate allows to mimic the semantics of calling all hooks one by one, with *Uni if the need it, without having access to *Uni.
func (e *Ev) Iterate(eventname string, stopfunc interface{}, params ...interface{}) {
	e.trigger(eventname, stopfunc, params...)
}

func (e *Ev) trigger(eventname string, stopfunc interface{}, params ...interface{}) {
	subscribed := all(e, eventname)
	hookname := hooknameize(eventname)
	var stopfunc_numin int
	if stopfunc != nil {
		s := reflect.TypeOf(stopfunc)
		if s.Kind() != reflect.Func {
			panic("Stopfunc is not a function.")
		}
		if s.NumOut() != 1 {
			panic("Stopfunc must have one return value.")
		}
		if s.Out(0) != reflect.TypeOf(false) {
			panic("Stopfunc must have a boolean return value.")
		}
		stopfunc_numin = s.NumIn()
	}
	for _, modname := range subscribed {
		hook_outp := []reflect.Value{}
		if !e.uni.Caller.Has("hooks", modname, hookname) {
			continue
		}
		var ret_rec interface{}
		if stopfunc != nil {
			ret_rec = func(i ...interface{}) {
				for i, v := range i {
					if v == nil {
						hook_outp = append(hook_outp, reflect.Zero(reflect.TypeOf(stopfunc).In(i)))
					} else {
						hook_outp = append(hook_outp, reflect.ValueOf(v))
					}
				}
			}
		}
		e.uni.Caller.Call("hooks", modname, hookname, ret_rec, params...)
		if stopfunc != nil {
			if stopfunc_numin != len(hook_outp) {
				panic(fmt.Sprintf("The number of return values of Hook %v of %v differs from the number of arguments of stopfunc.", hookname, modname))	// This sentence...
			}
			stopf := reflect.ValueOf(stopfunc)
			stopf_ret := stopf.Call(hook_outp)
			if stopf_ret[0].Interface().(bool) == true {
				break
			}
		}
	}
}

func NewEv(uni *Uni) *Ev {
	return &Ev{uni}
}

// Creates a hookname from access path.
// "content.insert" => "ContentInsert"
func hooknameize(s string) string {
	s = strings.Replace(s, ".", " ", -1)
	s = strings.Title(s)
	return strings.Replace(s, " ", "", -1)
}

var Convert = basic.Convert
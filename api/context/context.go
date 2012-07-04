// Context contains the type Uni. An instance of this type is passed to the modules when routing the control to them.
package context

import (
	"github.com/opesun/jsonp"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"net/http"
	"strings"
)

// General context for the application.
type Uni struct {
	Db    *mgo.Database
	W     http.ResponseWriter
	Req   *http.Request
	P     string					// Path string
	Paths []string 					// Path slice, contains the url (after the domain) splitted by "/"
	Opt   map[string]interface{}	// Freshest options from database.
	Dat   map[string]interface{} 	// General communication channel.
	Put   func(...interface{})   	// Just a convenienc function to allow fast output to http response.
	Root  string                 	// Absolute path of the application.
	Ev	  *Ev
	GetHook	func(string, string) func(*Uni) error 
}

// With the help of this type it's possible for the model to not have direct access to everything (*context.Uni), but still trigger events,
// which in turn will result in hooks (which will have access to everything) being called.
type Ev struct {
	uni 	*Uni
	Params []interface{}
}

// This is something I am cracking my head on, but for now it will be left out.
//
// type RunResult {
// 	HooksRan	int
// 	ErrorCount	int			// len of below slice, it's here for easier access.
// 	Errors		[]error
// }


// s : "content.insert", "content.blog.insert"
func (e *Ev) Trigger(s ...string) {
	for _, acc_path := range s {
		subscribed, has := jsonp.GetS(e.uni.Opt, acc_path)
		if has {
			for _, modname := range subscribed {
				hook := e.uni.GetHook(modname.(string), hooknameize(acc_path))
				if hook != nil {
					hook(e.uni)
				}
			}
		}
	}
	e.Params = make([]interface{}, 5)
}

func (e *Ev) Param(i ...interface{}) {
	for l, _ := range i {
		e.Params = append(e.Params, i[l])
	}
}

type Event interface{
	Param(i ...interface{})
	Trigger(s ...string)
}

func NewEv(uni *Uni) *Ev {
	return &Ev{uni, []interface{}{}}
}

// "insert.content" => "InsertContent"
func hooknameize(s string) string {
	s = strings.Replace(s, ".", " ", -1)
	s = strings.Title(s)
	return strings.Replace(s, " ", "", -1)
}

// Convert multiply nested bson.M-s to map[string]interface{}
// Written by Rog Peppe.
func Convert(x interface{}) interface{} {
	if x, ok := x.(bson.M); ok {
		for key, val := range x {
			x[key] = Convert(val)
		}
		return (map[string]interface{})(x)
	}
	return x
}

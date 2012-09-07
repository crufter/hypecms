// Context contains the type Uni. An instance of this type is passed to the modules when routing the control to them.
package context

import (
	"github.com/opesun/hypecms/model/basic"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo"
	"net/http"
	"strings"
	//"fmt"
)

// General context for the application.
type Uni struct {
	Db		*mgo.Database
	W		http.ResponseWriter
	Req		*http.Request
	secret	string						// Used for things like encryption/decryption. Basically a permanent random data.
	P		string						// Path string
	Paths	[]string 					// Path slice, contains the url (after the domain) splitted by "/"
	opt		string						// Original string representation of the option, if one needs a version which is guaranteedly untinkered.
	Opt		map[string]interface{}		// Freshest options from database.
	Dat		map[string]interface{} 		// General communication channel.
	Put		func(...interface{})   		// Just a convenience function to allow fast output to http response.
	Root	string                 		// Absolute path of the application.
	Ev		*Ev
	GetHook	func(string, string) interface{}
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

func (e *Ev) Param(params ...interface{}) {
	for i, _ := range params {
		e.Params = append(e.Params, params[i])
	}
}

// s : "content.insert", "content.blog.insert"
func (e *Ev) Trigger(eventname string, params ...interface{}) {
	e.Param(params...)
	e.TriggerAll(eventname)
}

func (e *Ev) TriggerAll(eventnames ...string) {
	//fmt.Println("Triggered event ", eventnames)
	for _, acc_path := range eventnames {
		subscribed, has := jsonp.GetS(e.uni.Opt, acc_path)
		if has {
			for _, modname := range subscribed {
				hook := e.uni.GetHook(modname.(string), hooknameize(acc_path))
				if hook != nil {
					//hook(e.uni)
				}
			}
		}
	}
	e.Params = make([]interface{}, 5)
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

var Convert = basic.Convert
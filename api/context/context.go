// Context contains the type Uni. An instance of this type is passed to the modules when routing the control to them.
package context

import (
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"net/http"
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

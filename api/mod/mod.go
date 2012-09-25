// This package gets around the lack of dynamic code loading in Go.
package mod

import(
	"github.com/opesun/hypecms/api/context"
	"reflect"
	"strings"
)

// Note: when
type dyn struct{
	Actions, Hooks, Views interface{}
}

var modules = map[string]dyn{
}

type Call struct{
	uni *context.Uni
}

func NewCall(uni *context.Uni) *Call {
	return &Call{uni}
}

func constr(what, module string) reflect.Value {
	what = strings.Title(what) // omg
	d, has := modules[module]
	if !has {
		return reflect.Value{} // fmt.Errorf("mod: No such module.")
	}
	// Seperatated for better readability.
	return reflect.ValueOf(d).FieldByName(what)		// When members of dyn were private they were not accessible.
}

func (c *Call) method(what, module, fname string) reflect.Value {
	constructor_i := constr(what, module)
	empty := reflect.Value{}
	if constructor_i == empty {
		return reflect.Value{}
	}
	constructor := reflect.ValueOf(constructor_i.Interface())
	constr_ret := constructor.Call([]reflect.Value{reflect.ValueOf(c.uni)})
	return constr_ret[0].MethodByName(fname)
}

// Maybe should return an error.
func (c *Call) Call(what, module, fname string, ret_reciever interface{}, params ...interface{}) {
	method := c.method(what, module, fname)
	empty := reflect.Value{}
	if method == empty {
		return
	}
	subj_in := []reflect.Value{}
	for _, v := range params {
		subj_in = append(subj_in, reflect.ValueOf(v))
	}
	subj_out := method.Call(subj_in)
	if ret_reciever != nil {
		reflect.ValueOf(ret_reciever).Call(subj_out)
	}
	return // nil
}

func (c *Call) Has(what, module, fname string) bool {
	method := c.method(what, module, fname)
	empty := reflect.Value{}
	if method == empty {
		return false
	}
	return method.Kind() == reflect.Func
}

// Method names
func (c *Call) Names(what, module string) []string {
	return []string{}
}

// Mathes signature
func (c *Call) Matches(what, module, fname string, i interface{}) bool  {
	return true
}


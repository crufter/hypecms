// This package gets around the lack of dynamic code loading in Go.
package mod

import(
	"github.com/opesun/hypecms/api/context"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

var empty = reflect.Value{}

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
		return empty // fmt.Errorf("mod: No such module.")
	}
	// Seperatated for better readability.
	return reflect.ValueOf(d).FieldByName(what)
}

func (c *Call) instance(what, module string) reflect.Value {
	constructor_i := constr(what, module)
	if constructor_i == empty {
		return empty
	}
	constructor := reflect.ValueOf(constructor_i.Interface())	// Isnt it unnecessary here? Test it.
	constr_ret := constructor.Call([]reflect.Value{reflect.ValueOf(c.uni)})
	return constr_ret[0]
}

func (c *Call) method(what, module, fname string) reflect.Value {
	inst := c.instance(what, module)
	if inst == empty {
		return empty
	}
	return inst.MethodByName(fname)
}

// Maybe should return an error.
func (c *Call) Call(what, module, fname string, ret_reciever interface{}, params ...interface{}) {
	method := c.method(what, module, fname)
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
	if method == empty {
		return false
	}
	return method.Kind() == reflect.Func
}

// Method names
func (c *Call) Names(what, module string) []string {
	inst := c.instance(what, module)
	t := reflect.TypeOf(inst.Interface())
	names := []string{}
	num := t.NumMethod()
	for i:=0;i<num;i++{
		mname := t.Method(i).Name
		r, _ := utf8.DecodeRuneInString(mname)
		if unicode.IsUpper(r) {
			names = append(names, mname)
		}
	}
	return names
}

// Mathes signature
func (c *Call) Matches(what, module, fname string, i interface{}) bool  {
	return true
}


// This package gets around the lack of dynamic code loading in Go.
package mod

import(
	"github.com/opesun/hypecms/frame/context"
	"reflect"
	"strings"
	"unicode"
	"fmt"
	"unicode/utf8"
)

var empty = reflect.Value{}

// Note: when
type dyn struct{
	Action, Hook, View interface{}
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
	if constructor_i.Interface() == nil {
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
func (c *Call) Call(what, module, fname string, ret_reciever interface{}, params ...interface{}) error {
	method := c.method(what, module, fname)
	if method == empty {
		return fmt.Errorf("mod: Can't find method.")
	}
	subj_in := []reflect.Value{}
	for _, v := range params {
		subj_in = append(subj_in, reflect.ValueOf(v))
	}
	subj_out := method.Call(subj_in)
	if ret_reciever != nil {
		reflect.ValueOf(ret_reciever).Call(subj_out)
	}
	return nil
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

func inputs(method reflect.Value) []reflect.Type {
	mtype := method.Type()
	in := mtype.NumIn()
	ret := []reflect.Type{}
	for i:=0;i<in;i++{
		ret = append(ret, mtype.In(i))
	}
	return ret
}

func (c *Call) Inputs(what, module, fname string) []reflect.Type {
	return inputs(c.method(what, module, fname))
}

func outputs(method reflect.Value) []reflect.Type {
	mtype := method.Type()
	out := mtype.NumOut()
	ret := []reflect.Type{}
	for i:=0;i<out;i++{
		ret = append(ret, mtype.Out(i))
	}
	return ret
}

func (c *Call) Outputs(what, module, fname string) []reflect.Type {
	return outputs(c.method(what, module, fname))
}

// Mathes signature
func (c *Call) Matches(what, module, fname string, i interface{}) bool  {
	return true
}


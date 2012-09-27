// This package gets around the lack of dynamic code loading in Go.
package mod

import(
	"github.com/opesun/hypecms/frame/context"
	"reflect"
	"unicode"
	"fmt"
	"unicode/utf8"
)

var empty = reflect.Value{}

type store map[string]reflect.Type

func (s store) register(modname string, a interface{}) {
	s[modname] = reflect.TypeOf(a)
}

var mods = store{}

type Call struct{
	uni *context.Uni
}

func NewCall(uni *context.Uni) *Call {
	return &Call{uni}
}

func emptyInstance(module string) reflect.Value {
	d, has := mods[module]
	if !has {
		return empty
	}
	return reflect.New(d)
}

func method(inst reflect.Value, fname string) reflect.Value {
	return inst.MethodByName(fname)
}

func (c *Call) init(empty_inst reflect.Value) {
	initfunc := method(empty_inst, "Init")
	initfunc.Call([]reflect.Value{reflect.ValueOf(c.uni)})	// Check for error maybe?
}

// Maybe should return an error.
func (c *Call) Call(module, fname string, ret_reciever interface{}, params ...interface{}) error {
	e_inst := emptyInstance(module)
	c.init(e_inst)
	meth := method(e_inst, fname)	// Hehehe.
	if meth == empty {
		return fmt.Errorf("mod: Can't find method.")
	}
	subj_in := []reflect.Value{}
	for _, v := range params {
		subj_in = append(subj_in, reflect.ValueOf(v))
	}
	subj_out := meth.Call(subj_in)
	if ret_reciever != nil {
		reflect.ValueOf(ret_reciever).Call(subj_out)
	}
	return nil
}

func (c *Call) Has(module, fname string) bool {
	e_inst := emptyInstance(module)
	meth := method(e_inst, fname)
	if meth == empty {
		return false
	}
	return meth.Kind() == reflect.Func
}

// Method names
func (c *Call) Names(module string) []string {
	inst := emptyInstance(module)
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

func inputs(meth reflect.Value) []reflect.Type {
	mtype := meth.Type()
	in := mtype.NumIn()
	ret := []reflect.Type{}
	for i:=0;i<in;i++{
		ret = append(ret, mtype.In(i))
	}
	return ret
}

func (c *Call) Inputs(module, fname string) []reflect.Type {
	e_inst := emptyInstance(module)
	meth := method(e_inst, fname)
	return inputs(meth)
}

func outputs(meth reflect.Value) []reflect.Type {
	mtype := meth.Type()
	out := mtype.NumOut()
	ret := []reflect.Type{}
	for i:=0;i<out;i++{
		ret = append(ret, mtype.Out(i))
	}
	return ret
}

func (c *Call) Outputs(module, fname string) []reflect.Type {
	e_inst := emptyInstance(module)
	meth := method(e_inst, fname)
	return inputs(meth)
}

// Mathes signature
func (c *Call) Matches(module, fname string, i interface{}) bool  {
	return true
}


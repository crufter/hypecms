// This package gets around the lack of dynamic code loading in Go.
package mod

import(
	iface "github.com/opesun/hypecms/frame/interfaces"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var empty = reflect.Value{}

type store map[string]reflect.Type

func (s store) register(modname string, a interface{}) {
	s[modname] = reflect.TypeOf(a)
}

var mods = store{}

type Module struct {
	name string
}

func NewModule(s string) iface.Module {
	return &Module{s}
}

func (m *Module) Exists() bool {
	_, has := mods[m.name]
	return has
}

func (m *Module) Instance() iface.Instance {
	return Instance(reflect.New(mods[m.name]))
}

type Instance reflect.Value

func (i Instance) HasMethod(fname string) bool {
	return reflect.Value(i).MethodByName(fname).Kind() == reflect.Func
}

func (i Instance) MethodNames() []string {
	t := reflect.TypeOf(reflect.Value(i).Interface())
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

func(i Instance) Method(s string) iface.Method {
	return Method(reflect.Value(i).MethodByName(s))
}

type Method reflect.Value

func (me Method) Call(ret_reciever interface{}, params ...interface{}) error {
	subj_in := []reflect.Value{}
	for _, v := range params {
		subj_in = append(subj_in, reflect.ValueOf(v))
	}
	subj_out := reflect.Value(me).Call(subj_in)
	if ret_reciever != nil {
		reflect.ValueOf(ret_reciever).Call(subj_out)
	}
	return nil
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

func (me Method) InputTypes() []reflect.Type {
	return inputs(reflect.Value(me))
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

func (me Method) OutputTypes() []reflect.Type {
	return outputs(reflect.Value(me))
}

// Mathes signature
func (me Method) Matches(i interface{}) bool  {
	return true
}


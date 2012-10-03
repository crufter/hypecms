package main

import(
	"reflect"
	iface "github.com/opesun/hypecms/frame/interfaces"
)

type Analyzer struct {
	m iface.Method
}

func NewAnalyzer(method iface.Method) *Analyzer {
	return &Analyzer{method}
}

func (a *Analyzer) ArgCount() int {
	return len(m.Inputs)
}

func (a *Analyzer) FilterCount() int {
	inptypes := m.Inputs()
	c := 0
	var i iface.Filter{}
	ft := reflect.TypeOf(i)
	for _, v := range inptypes {
		if v.Implements(ft) {
			c++	// lol.
		}
	}
	return c
}
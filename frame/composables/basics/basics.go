package basics

import(
	//"labix.org/v2/mgo"
	iface "github.com/opesun/hypecms/frame/interfaces"
)

type Basics struct{
}

func (b *Basics) Get(a iface.Filter) ([]interface{}, error) {
	return nil, nil
}

func (b *Basics) Post(a iface.Filter) ([]interface{}, error) {
	return nil, nil
}

func (b *Basics) Put(a iface.Filter) {
}
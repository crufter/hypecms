package basics

import(
	//"labix.org/v2/mgo"
	iface "github.com/opesun/hypecms/frame/interfaces"
)

type Basics struct{
}

func (b *Basics) Get(a iface.Filter) ([]interface{}, error) {
	return a.Find()
}

func (b *Basics) Insert(a iface.Filter, data map[string]interface{}) (error) {
	return a.Insert(data)
}

func (b *Basics) Update(a iface.Filter, data map[string]interface{}) error {
	return a.Update(data)
}

func (b *Basics) Remove(a iface.Filter) error {
	return a.Remove()
}
package qbgutils

import "google.golang.org/appengine/datastore"

type Plugin interface {
	Init(typeName string)
	Ancestor(ancestor *datastore.Key)
	KeysOnly()
	Start(cur datastore.Cursor)
	Offset(offset int)
	Limit(limit int)
	Filter(name, op string, value interface{})
	Asc(name string)
	Desc(name string)
}

type Plugger interface {
	Plugin() Plugin
}

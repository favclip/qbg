package qbgutils

import "google.golang.org/appengine/datastore"

// Plugin supply hook point for query constructions.
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

// Plugger supply Plugin component.
type Plugger interface {
	Plugin() Plugin
}

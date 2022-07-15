package b

import "time"

type Sample struct {
	ID        int64     `datastore:"-" goon:"id"`
	CreatedAt time.Time `datastore:",noindex"`
}

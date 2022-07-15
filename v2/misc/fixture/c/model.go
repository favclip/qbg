//go:generate qbg -output model_query.go .

package c

import "time"

// +qbg
type Sample struct {
	ID        int64     `goon:"id"`
	CreatedAt time.Time `datastore:",noindex"`
}

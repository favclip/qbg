//go:generate qbg -output model_query.go .

package c

import "time"

// +qbg
type Sample struct {
	ID        int64       `goon:"id"`
	FooUrl    string      `datastore:"FooURL,noindex"`
	Start     time.Ticker `qbg:"StartAt"`
	Limit     int
	CreatedAt time.Time `datastore:",noindex"`
}

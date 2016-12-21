//go:generate qbg -output model_query.go .
package f

// +qbg
type Sample struct {
	Kind string `goon:"kind,sample_kind"`
	Foo  string
}

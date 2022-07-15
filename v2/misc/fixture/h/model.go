//go:generate qbg -output model_query.go -inlineinterfaces .
package h

// +qbg
type Sample struct {
	Kind string `goon:"kind,sample_kind"`
	Foo  string
}

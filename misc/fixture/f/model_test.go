package f

import (
	"testing"

	"github.com/mjibson/goon"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

func TestGoonKind(t *testing.T) {
	opts := &aetest.Options{
		StronglyConsistentDatastore: true,
	}
	inst, err := aetest.NewInstance(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	c := appengine.NewContext(req)

	g := goon.FromContext(c)

	entity1 := &Sample{
		Foo: "foo1",
	}
	key1, err := g.Put(entity1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("key1: %s", key1.String())

	entity2 := &Sample{
		Kind: "sample_specified",
		Foo:  "foo1",
	}
	key2, err := g.Put(entity2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("key2: %s", key2.String())

	cnt1, err := NewSampleQueryBuilder().Query().Count(c)
	if err != nil {
		t.Fatal(err)
	}
	if cnt1 != 1 {
		t.Errorf("unexpected: %v", cnt1)
	}

	cnt2, err := NewSampleQueryBuilderWithKind(entity2.Kind).Query().Count(c)
	if err != nil {
		t.Fatal(err)
	}
	if cnt2 != 1 {
		t.Errorf("unexpected: %v", cnt2)
	}
}

package e

import "testing"

func TestPlugin(t *testing.T) {
	qb := NewSampleQueryBuilder()
	qb.Foo.Equal("test")
	qb.Foo.Desc()
	qb.KeysOnly()
	qb.Offset(0)
	qb.Limit(10)

	mp, ok := qb.Plugin().(*MemcacheQueryPlugin)
	if !ok {
		t.Fatal("Plugin is not MemcacheQueryPlugin")
	}

	if str := mp.QueryString(); str != `k=Sample:?Foo="test":D=Foo:!k:!o=0:!l=10` {
		t.Fatalf("unexpected: %s", str)
	}
}

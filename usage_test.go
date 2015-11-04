package qbg

import (
	"testing"

	"github.com/favclip/qbg/misc/fixture/a"
	"github.com/favclip/qbg/misc/fixture/e"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

func TestBasicUsage1(t *testing.T) {
	c, closer, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	src := &a.Sample{"Foo!"}
	key := datastore.NewIncompleteKey(c, "Sample", nil)
	_, err = datastore.Put(c, key, src)
	if err != nil {
		t.Fatal(err)
	}

	builder := a.NewSampleQueryBuilder()
	builder.Foo.GreaterThan("Foo")
	iter := builder.Query().Run(c)
	for {
		src := &a.Sample{}
		key, err = iter.Next(src)
		if err == datastore.Done {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		t.Logf("key: %#v, entity: %#v", key, src)
	}
}

func TestBasicUsage2(t *testing.T) {
	c, closer, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	src := &a.Sample{"Foo!"}
	key := datastore.NewIncompleteKey(c, "Sample", nil)
	_, err = datastore.Put(c, key, src)
	if err != nil {
		t.Fatal(err)
	}

	builder := a.NewSampleQueryBuilder()
	builder.Foo.GreaterThan("Foo").KeysOnly().Limit(3)
	iter := builder.Query().Run(c)
	for {
		key, err = iter.Next(nil)
		if err == datastore.Done {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		src := &a.Sample{}
		err = datastore.Get(c, key, src)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("key: %#v, entity: %#v", key, src)
	}
}

func TestPluginUsage(t *testing.T) {
	c, closer, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	parentKey := datastore.NewKey(c, "Test", "T", 0, nil)

	{
		// put for Datastore
		src := &e.Sample{"Foo!"}
		key := datastore.NewIncompleteKey(c, "Sample", parentKey)
		_, err = datastore.Put(c, key, src)
		if err != nil {
			t.Fatal(err)
		}
		_, err = memcache.Increment(c, "counter", 1, 0) // shift memcache counter
		if err != nil {
			t.Fatal(err)
		}
	}

	builder := e.NewSampleQueryBuilder()
	mp, ok := builder.Plugin().(*e.MemcacheQueryPlugin)
	if !ok {
		t.Fatal("Plugin is not MemcacheQueryPlugin")
	}
	memcacheCounter, err := memcache.Increment(c, "counter", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	mp.AddCounter(memcacheCounter)

	builder.Ancestor(parentKey).Foo.GreaterThan("Fo").Limit(3)
	if str := mp.QueryString(); str != `k=Sample:1:!a=/Test,T:?Foo>"Fo":!l=3` {
		t.Fatalf("unexpected: %s", str)
	}

	{
		// get from Memcache
		var list []*e.Sample
		_, err = memcache.Gob.Get(c, mp.QueryString(), &list)
		if err == memcache.ErrCacheMiss {
			// continue
		} else if err != nil {
			t.Fatal(err)
		} else {
			t.Log(list)
			return
		}
	}

	{
		// get from Datastore
		iter := builder.Query().Run(c)
		var list []*e.Sample
		for {
			src := &e.Sample{}
			_, err = iter.Next(src)
			if err == datastore.Done {
				break
			} else if err != nil {
				t.Fatal(err)
			}
			list = append(list, src)
		}

		// put for Memcache
		err = memcache.Gob.Set(c, &memcache.Item{
			Key:    mp.QueryString(),
			Object: list,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 1 {
			t.Fatalf("unexpected %d", len(list))
		}
	}

	{
		// get from Memcache
		var list []*e.Sample
		_, err = memcache.Gob.Get(c, mp.QueryString(), &list)
		if err == memcache.ErrCacheMiss {
			t.Fatalf("query result is not in memcache")
		} else if err != nil {
			t.Fatal(err)
		} else {
			t.Log(list)
		}
	}
}

# qbg

`Query Builder Generator`.

## Description

`qbg` can generate type safe wrapper for appengine datastore.

If you use string literal for building query, You shoots your foot when you do typo.
qbg generate appengine datastore wrapper code. It is type safe. Your mistake will be found by go compiler.

```
type User struct {
	Name	string
}
```

```
user := &User {
	"go-chan",
}
builder := NewUserQueryBuilder()
builder.Name.Equal("go-chan")
cnt, err := builder.Query().Count(c)
```

### Example

[from](https://github.com/favclip/qbg/blob/master/misc/fixture/a/model.go):

```
type Sample struct {
	Foo string
}
```

[to](https://github.com/favclip/qbg/blob/master/misc/fixture/a/model_query.go):

```
// generated!
// for Sample
type SampleQueryBuilder struct {
	q      *datastore.Query
	plugin qbgutils.Plugin
	Foo    *SampleQueryProperty
}

type SampleQueryProperty struct {
	bldr *SampleQueryBuilder
	name string
}
```

usage:

```
src := &Sample{"Foo!"}

builder := NewSampleQueryBuilder() // generated!
builder.Foo.GreaterThanOrEqual("Foo")
cnt, err := builder.Query().Count(c)
```

[other example](https://github.com/favclip/qbg/blob/master/usage_test.go).

### With `go generate`

```
$ ls -la .
total 8
drwxr-xr-x@ 3 vvakame  staff  102 10 13 17:39 .
drwxr-xr-x@ 7 vvakame  staff  238  8 14 18:26 ..
-rw-r--r--@ 1 vvakame  staff  178  8 14 18:26 model.go
$ cat model.go
//go:generate qbg -output model_query.go .

package c

import "time"

// +qbg
type Sample struct {
	ID        int64     `goon:"id"`
	CreatedAt time.Time `datastore:",noindex"`
}
$ go generate
$ ls -la .
total 16
drwxr-xr-x@ 4 vvakame  staff   136 10 13 17:40 .
drwxr-xr-x@ 7 vvakame  staff   238  8 14 18:26 ..
-rw-r--r--@ 1 vvakame  staff   178  8 14 18:26 model.go
-rw-r--r--  1 vvakame  staff  3709 10 13 17:40 model_query.go
```

### Recommend

Please use with [goon](https://github.com/mjibson/goon).

## Installation

```
$ go get -u github.com/favclip/qbg/cmd/qbg
$ qbg
Usage of qbg:
	qbg [flags] [directory]
	qbg [flags] files... # Must be a single package
Flags:
  -output="": output file name; default srcdir/<type>_query.go
  -type="": comma-separated list of type names; must be set
```

## Command sample

Model with type specific option.

```
$ cat misc/fixture/a/model.go
package a

// test for basic struct definition

type Sample struct {
	Foo string
}
$ qbg -type Sample -output misc/fixture/a/model_query.go misc/fixture/a
```

Model with tagged comment.

```
$ cat misc/fixture/c/model.go
//go:generate qbg -output model_query.go .

package c

import "time"

// +qbg
type Sample struct {
	ID        int64     `goon:"id"`
	CreatedAt time.Time `datastore:",noindex"`
}
$ qbg -output misc/fixture/d/model_query.go misc/fixture/d
```

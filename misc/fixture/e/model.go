package e

import (
	"bytes"
	"fmt"

	"github.com/favclip/qbg/qbgutils"
	"google.golang.org/appengine/datastore"
)

// +qbg
type Sample struct {
	Foo string
}

func (bldr *SampleQueryBuilder) Plugin() qbgutils.Plugin {
	if bldr.plugin == nil {
		bldr.plugin = &MemcacheQueryPlugin{}
	}
	return bldr.plugin
}

type MemcacheQueryPlugin struct {
	buf bytes.Buffer
}

func (p *MemcacheQueryPlugin) AddCounter(counter uint64) {
	p.buf.WriteString(fmt.Sprintf(":%d", counter))
}

func (p *MemcacheQueryPlugin) Init(typeName string) {
	p.buf.WriteString(fmt.Sprintf("k=%s", typeName))
}

func (p *MemcacheQueryPlugin) Ancestor(ancestor *datastore.Key) {
	p.buf.WriteString(fmt.Sprintf(":!a=%s", ancestor.String()))
}

func (p *MemcacheQueryPlugin) KeysOnly() {
	p.buf.WriteString(":!k")
}

func (p *MemcacheQueryPlugin) Start(cur datastore.Cursor) {
	p.buf.WriteString(fmt.Sprintf(":!s=%s", cur.String()))
}

func (p *MemcacheQueryPlugin) Offset(offset int) {
	p.buf.WriteString(fmt.Sprintf(":!o=%d", offset))
}

func (p *MemcacheQueryPlugin) Limit(limit int) {
	p.buf.WriteString(fmt.Sprintf(":!l=%d", limit))
}

func (p *MemcacheQueryPlugin) Filter(name, op string, value interface{}) {
	p.buf.WriteString(fmt.Sprintf(":?%s%s%#v", name, op, value))
}

func (p *MemcacheQueryPlugin) Asc(name string) {
	p.buf.WriteString(fmt.Sprintf(":A=%s", name))
}

func (p *MemcacheQueryPlugin) Desc(name string) {
	p.buf.WriteString(fmt.Sprintf(":D=%s", name))
}

func (p *MemcacheQueryPlugin) QueryString() string {
	return p.buf.String()
}

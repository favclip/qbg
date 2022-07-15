package qbg

const pluginTemplate = `
// Plugin supply hook point for query constructions.
type Plugin interface {
	Init(typeName string)
	Ancestor(ancestor {{.DSKeyType}})
	KeysOnly()
	Start(cur datastore.Cursor)
	Offset(offset int)
	Limit(limit int)
	Filter(name, op string, value interface{})
	Asc(name string)
	Desc(name string)
}

// Plugger supply Plugin component.
type Plugger interface {
	Plugin() Plugin
}
`

const structTemplate = `
{{$st := .}}
// {{.Name}}QueryBuilder build query for {{.Name}}.
type {{.Name}}QueryBuilder struct {
	q {{.DSQueryType}}
	plugin {{.PluginType}}
	{{- range $idx, $f := .Fields}}
		{{$f.Tag.PropertyNameAlter}} *{{$st.Name}}QueryProperty
	{{- end}}
}

// {{.Name}}QueryProperty has property information for {{.Name}}QueryBuilder.
type {{.Name}}QueryProperty struct {
	bldr *{{.Name}}QueryBuilder
	name string
}

// {{.NewWord}}{{.SimpleName}}QueryBuilder create new {{.SimpleName}}QueryBuilder.
func {{.NewWord}}{{.SimpleName}}QueryBuilder({{if .UserDatastoreWrapper}}client datastore.Client,{{end}}) *{{.Name}}QueryBuilder {
	return {{.NewWord}}{{.SimpleName}}QueryBuilderWithKind({{if .UserDatastoreWrapper}}client,{{end}}"{{.Kind}}")
}

// {{.NewWord}}{{.SimpleName}}QueryBuilderWithKind create new {{.SimpleName}}QueryBuilder with specific kind.
func {{.NewWord}}{{.SimpleName}}QueryBuilderWithKind({{if .UserDatastoreWrapper}}client datastore.Client,{{end}}kind string) *{{.Name}}QueryBuilder {
	q := {{.DSNewQuery}}(kind)
	bldr := &{{.Name}}QueryBuilder{q:q}
	{{- range $idx, $f := .Fields}}
		bldr.{{$f.Tag.PropertyNameAlter}}= &{{$st.Name}}QueryProperty{
		bldr : bldr,
		name : {{if $f.Tag.ID}} "__key__" {{else}} "{{$f.Tag.Name}}" {{end}},
	}
	{{- end}}

	if plugger, ok := interface{}(bldr).({{.PluggerType}}); ok {
		bldr.plugin = plugger.Plugin()
		bldr.plugin.Init("{{.SimpleName}}")
	}

	return bldr
}

// Ancestor sets parent key to ancestor query.
func (bldr *{{.Name}}QueryBuilder) Ancestor(parentKey {{.DSKeyType}}) *{{.Name}}QueryBuilder {
	bldr.q = bldr.q.Ancestor(parentKey)
	if bldr.plugin != nil {
		bldr.plugin.Ancestor(parentKey)
	}
	return bldr
}

// KeysOnly sets keys only option to query.
func (bldr *{{.Name}}QueryBuilder) KeysOnly() *{{.Name}}QueryBuilder {
	bldr.q = bldr.q.KeysOnly()
	if bldr.plugin != nil {
		bldr.plugin.KeysOnly()
	}
	return bldr
}

// Start setup to query.
func (bldr *{{.Name}}QueryBuilder) Start(cur datastore.Cursor) *{{.Name}}QueryBuilder {
	bldr.q = bldr.q.Start(cur)
	if bldr.plugin != nil {
		bldr.plugin.Start(cur)
	}
	return bldr
}

// Offset setup to query.
func (bldr *{{.Name}}QueryBuilder) Offset(offset int) *{{.Name}}QueryBuilder {
	bldr.q = bldr.q.Offset(offset)
	if bldr.plugin != nil {
		bldr.plugin.Offset(offset)
	}
	return bldr
}

// Limit setup to query.
func (bldr *{{.Name}}QueryBuilder) Limit(limit int) *{{.Name}}QueryBuilder {
	bldr.q = bldr.q.Limit(limit)
	if bldr.plugin != nil {
		bldr.plugin.Limit(limit)
	}
	return bldr
}

// Query returns *datastore.Query.
func (bldr *{{.Name}}QueryBuilder) Query() {{.DSQueryType}} {
	return bldr.q
}

// Filter with op & value.
func (p *{{.Name}}QueryProperty) Filter(op string, value interface{}) *{{.Name}}QueryBuilder {
	switch op {
		case "<=":
			p.LessThanOrEqual(value)
		case ">=":
			p.GreaterThanOrEqual(value)
		case "<":
			p.LessThan(value)
		case ">":
			p.GreaterThan(value)
		case "=":
			p.Equal(value)
		default:
			p.bldr.q = p.bldr.q.Filter(p.name + " " + op, value) // error raised by native query
	}
	if p.bldr.plugin != nil {
		p.bldr.plugin.Filter(p.name, op, value)
	}
	return p.bldr
}

// LessThanOrEqual filter with value.
func (p *{{.Name}}QueryProperty) LessThanOrEqual(value interface{}) *{{.Name}}QueryBuilder {
	p.bldr.q = p.bldr.q.Filter(p.name + " <=", value)
	if p.bldr.plugin != nil {
		p.bldr.plugin.Filter(p.name,"<=", value)
	}
	return p.bldr
}

// GreaterThanOrEqual filter with value.
func (p *{{.Name}}QueryProperty) GreaterThanOrEqual(value interface{}) *{{.Name}}QueryBuilder {
	p.bldr.q = p.bldr.q.Filter(p.name +  " >=", value)
	if p.bldr.plugin != nil {
		p.bldr.plugin.Filter(p.name,">=", value)
	}
	return p.bldr
}

// LessThan filter with value.
func (p *{{.Name}}QueryProperty) LessThan(value interface{}) *{{.Name}}QueryBuilder {
	p.bldr.q = p.bldr.q.Filter(p.name + " <", value)
	if p.bldr.plugin != nil {
		p.bldr.plugin.Filter(p.name,"<", value)
	}
	return p.bldr
}

// GreaterThan filter with value.
func (p *{{.Name}}QueryProperty) GreaterThan(value interface{}) *{{.Name}}QueryBuilder {
	p.bldr.q = p.bldr.q.Filter(p.name + " >", value)
	if p.bldr.plugin != nil {
		p.bldr.plugin.Filter(p.name,">", value)
	}
	return p.bldr
}

// Equal filter with value.
func (p *{{.Name}}QueryProperty) Equal(value interface{}) *{{.Name}}QueryBuilder {
	p.bldr.q = p.bldr.q.Filter(p.name + " =", value)
	if p.bldr.plugin != nil {
		p.bldr.plugin.Filter(p.name,"=", value)
	}
	return p.bldr
}

// Asc order.
func (p *{{.Name}}QueryProperty) Asc() *{{.Name}}QueryBuilder {
	p.bldr.q = p.bldr.q.Order(p.name)
	if p.bldr.plugin != nil {
		p.bldr.plugin.Asc(p.name)
	}
	return p.bldr
}

// Desc order.
func (p *{{.Name}}QueryProperty) Desc() *{{.Name}}QueryBuilder {
	p.bldr.q = p.bldr.q.Order("-" + p.name)
	if p.bldr.plugin != nil {
		p.bldr.plugin.Desc(p.name)
	}
	return p.bldr
}

`

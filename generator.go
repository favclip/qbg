package qbg

import (
	"reflect"
	"strings"

	"github.com/favclip/genbase"
)

// BuildSource represents source code of assembling..
type BuildSource struct {
	g         *genbase.Generator
	pkg       *genbase.PackageInfo
	typeInfos genbase.TypeInfos

	Structs []*BuildStruct
}

// BuildStruct represents struct of assembling..
type BuildStruct struct {
	parent   *BuildSource
	typeInfo *genbase.TypeInfo

	Fields []*BuildField
}

// BuildField represents field of BuildStruct.
type BuildField struct {
	parent    *BuildStruct
	fieldInfo *genbase.FieldInfo

	Name string
	Tag  *BuildTag
}

// BuildTag represents tag of BuildField.
type BuildTag struct {
	field *BuildField

	Kind              string // e.g. `goon:"kind,FooKind"`
	Name              string
	PropertyNameAlter string // e.g. `qbg:"StartAt"`
	ID                bool
	Ignore            bool // e.g. Secret string `datastore:"-"`
	NoIndex           bool // e.g. no index `datastore:",noindex"`
}

// Parse construct *BuildSource from package & type information.
func Parse(pkg *genbase.PackageInfo, typeInfos genbase.TypeInfos) (*BuildSource, error) {
	bu := &BuildSource{
		g:         genbase.NewGenerator(pkg),
		pkg:       pkg,
		typeInfos: typeInfos,
	}

	bu.g.AddImport("google.golang.org/appengine/datastore", "")
	bu.g.AddImport("github.com/favclip/qbg/qbgutils", "")

	for _, typeInfo := range typeInfos {
		err := bu.parseStruct(typeInfo)
		if err != nil {
			return nil, err
		}
	}

	return bu, nil
}

func (b *BuildSource) parseStruct(typeInfo *genbase.TypeInfo) error {
	structType, err := typeInfo.StructType()
	if err != nil {
		return err
	}

	st := &BuildStruct{
		parent:   b,
		typeInfo: typeInfo,
	}

	for _, fieldInfo := range structType.FieldInfos() {
		if len := len(fieldInfo.Names); len == 0 {
			// embedded struct in outer struct or multiply field declarations
			// https://play.golang.org/p/bcxbdiMyP4
			continue
		}

		for _, nameIdent := range fieldInfo.Names {
			err := b.parseField(st, typeInfo, fieldInfo, nameIdent.Name)
			if err != nil {
				return err
			}
		}
	}

	b.Structs = append(b.Structs, st)

	return nil
}

func (b *BuildSource) parseField(st *BuildStruct, typeInfo *genbase.TypeInfo, fieldInfo *genbase.FieldInfo, name string) error {
	field := &BuildField{
		parent:    st,
		fieldInfo: fieldInfo,
		Name:      name,
	}
	st.Fields = append(st.Fields, field)

	tag := &BuildTag{
		field: field,
		Name:  name,
	}
	field.Tag = tag

	if fieldInfo.Tag != nil {
		// remove back quote
		tagBody := fieldInfo.Tag.Value[1 : len(fieldInfo.Tag.Value)-1]
		tagKeys := genbase.GetKeys(tagBody)
		structTag := reflect.StructTag(tagBody)
		for _, key := range tagKeys {
			if key == "datastore" {
				tagText := structTag.Get("datastore")
				if tagText == "-" && !tag.ID {
					tag.Ignore = true
					continue
				}
				if idx := strings.Index(tagText, ","); idx == -1 {
					tag.Name = tagText
				} else {
					for idx != -1 || tagText != "" {
						value := tagText
						if idx != -1 {
							value = tagText[:idx]
							tagText = tagText[idx+1:]
						} else {
							tagText = tagText[len(value):]
						}
						idx = strings.Index(tagText, ",")

						if value == "noindex" {
							tag.NoIndex = true
						} else if value != "" {
							tag.Name = value
						}
					}
				}
			} else if key == "goon" {
				tagText := structTag.Get("goon")
				if tagText == "id" {
					tag.ID = true
					tag.Ignore = false
				} else if strings.HasPrefix(tagText, "kind,") {
					tag.Kind = tagText[5:]
				}
			} else if key == "qbg" {
				tag.PropertyNameAlter = structTag.Get("qbg")
			}
		}
	}
	if tag.PropertyNameAlter == "" {
		switch field.Name {
		case "Ancestor", "KeysOnly", "Start", "Offset", "Limit", "Query":
			tag.PropertyNameAlter = field.Name + "Property"
		default:
			tag.PropertyNameAlter = field.Name
		}
	}

	return nil
}

// Emit generate wrapper code.
func (b *BuildSource) Emit(args *[]string) ([]byte, error) {
	b.g.PrintHeader("qbg", args)

	for _, st := range b.Structs {
		err := st.emit(b.g)
		if err != nil {
			return nil, err
		}
	}

	return b.g.Format()
}

func (st *BuildStruct) emit(g *genbase.Generator) error {
	g.Printf("// %[1]sQueryBuilder build query for %[1]s.\n", st.Name())

	// generate FooQueryBuilder struct from Foo struct
	g.Printf("type %sQueryBuilder struct {\n", st.Name())
	g.Printf("q *datastore.Query\n")
	g.Printf("plugin qbgutils.Plugin\n")

	for _, field := range st.Fields {
		if field.Tag.Ignore {
			continue
		}

		g.Printf("%[1]s *%[2]sQueryProperty\n", field.Tag.PropertyNameAlter, st.Name())
	}
	g.Printf("}\n\n")

	// generate property info
	g.Printf(`
			// %[1]sQueryProperty has property information for %[1]sQueryBuilder.
			type %[1]sQueryProperty struct {
				bldr *%[1]sQueryBuilder
				name string
			}

			`, st.Name())

	// generate new query builder factory function
	g.Printf("// New%[1]sQueryBuilder create new %[1]sQueryBuilder.\n", st.Name())
	g.Printf("func New%[1]sQueryBuilder() *%[1]sQueryBuilder {\n", st.Name())
	g.Printf("return New%[1]sQueryBuilderWith(\"%[2]s\")\n", st.Name(), st.Kind())
	g.Printf("}\n\n")

	g.Printf("// New%[1]sQueryBuilderWith create new %[1]sQueryBuilder with specified kind.\n", st.Name())
	g.Printf("func New%[1]sQueryBuilderWith(kind string) *%[1]sQueryBuilder {\n", st.Name())
	g.Printf("q := datastore.NewQuery(kind)\n")
	g.Printf("bldr := &%[1]sQueryBuilder{q:q}\n", st.Name())
	for _, field := range st.Fields {
		if field.Tag.Ignore {
			continue
		}
		name := field.Tag.Name
		if field.Tag.ID {
			name = "__key__"
		}
		g.Printf(`bldr.%[2]s= &%[1]sQueryProperty{
						bldr : bldr,
						name: "%[3]s",
					}
			`, st.Name(), field.Tag.PropertyNameAlter, name)
	}
	g.Printf(`
			if plugger, ok := interface{}(bldr).(qbgutils.Plugger); ok {
				bldr.plugin = plugger.Plugin()
				bldr.plugin.Init("%[1]s")
			}
			`, st.Name())
	g.Printf("return bldr\n")
	g.Printf("}\n\n")

	// generate methods
	g.Printf(`
			// Ancestor sets parent key to ancestor query.
			func (bldr *%[1]sQueryBuilder) Ancestor(parentKey *datastore.Key) *%[1]sQueryBuilder {
				bldr.q = bldr.q.Ancestor(parentKey)
				if bldr.plugin != nil {
					bldr.plugin.Ancestor(parentKey)
				}
				return bldr
			}

			// KeysOnly sets keys only option to query.
			func (bldr *%[1]sQueryBuilder) KeysOnly() *%[1]sQueryBuilder {
				bldr.q = bldr.q.KeysOnly()
				if bldr.plugin != nil {
					bldr.plugin.KeysOnly()
				}
				return bldr
			}

			// Start setup to query.
			func (bldr *%[1]sQueryBuilder) Start(cur datastore.Cursor) *%[1]sQueryBuilder {
				bldr.q = bldr.q.Start(cur)
				if bldr.plugin != nil {
					bldr.plugin.Start(cur)
				}
				return bldr
			}

			// Offset setupto query.
			func (bldr *%[1]sQueryBuilder) Offset(offset int) *%[1]sQueryBuilder {
				bldr.q = bldr.q.Offset(offset)
				if bldr.plugin != nil {
					bldr.plugin.Offset(offset)
				}
				return bldr
			}

			// Limit setup to query.
			func (bldr *%[1]sQueryBuilder) Limit(limit int) *%[1]sQueryBuilder {
				bldr.q = bldr.q.Limit(limit)
				if bldr.plugin != nil {
					bldr.plugin.Limit(limit)
				}
				return bldr
			}

			// Query returns *datastore.Query.
			func (bldr *%[1]sQueryBuilder) Query() *datastore.Query {
				return bldr.q
			}

			// Filter with op & value.
			func (p *%[1]sQueryProperty) Filter(op string, value interface{}) *%[1]sQueryBuilder {
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
			func (p *%[1]sQueryProperty) LessThanOrEqual(value interface{}) *%[1]sQueryBuilder {
				p.bldr.q = p.bldr.q.Filter(p.name + " <=", value)
				if p.bldr.plugin != nil {
					p.bldr.plugin.Filter(p.name,"<=", value)
				}
				return p.bldr
			}

			// GreaterThanOrEqual filter with value.
			func (p *%[1]sQueryProperty) GreaterThanOrEqual(value interface{}) *%[1]sQueryBuilder {
				p.bldr.q = p.bldr.q.Filter(p.name +  " >=", value)
				if p.bldr.plugin != nil {
					p.bldr.plugin.Filter(p.name,">=", value)
				}
				return p.bldr
			}

			// LessThan filter with value.
			func (p *%[1]sQueryProperty) LessThan(value interface{}) *%[1]sQueryBuilder {
				p.bldr.q = p.bldr.q.Filter(p.name + " <", value)
				if p.bldr.plugin != nil {
					p.bldr.plugin.Filter(p.name,"<", value)
				}
				return p.bldr
			}

			// GreaterThan filter with value.
			func (p *%[1]sQueryProperty) GreaterThan(value interface{}) *%[1]sQueryBuilder {
				p.bldr.q = p.bldr.q.Filter(p.name + " >", value)
				if p.bldr.plugin != nil {
					p.bldr.plugin.Filter(p.name,">", value)
				}
				return p.bldr
			}

			// Equal filter with value.
			func (p *%[1]sQueryProperty) Equal(value interface{}) *%[1]sQueryBuilder {
				p.bldr.q = p.bldr.q.Filter(p.name + " =", value)
				if p.bldr.plugin != nil {
					p.bldr.plugin.Filter(p.name,"=", value)
				}
				return p.bldr
			}

			// Asc order.
			func (p *%[1]sQueryProperty) Asc() *%[1]sQueryBuilder {
				p.bldr.q = p.bldr.q.Order(p.name)
				if p.bldr.plugin != nil {
					p.bldr.plugin.Asc(p.name)
				}
				return p.bldr
			}

			// Desc order.
			func (p *%[1]sQueryProperty) Desc() *%[1]sQueryBuilder {
				p.bldr.q = p.bldr.q.Order("-" + p.name)
				if p.bldr.plugin != nil {
					p.bldr.plugin.Desc(p.name)
				}
				return p.bldr
			}
		`, st.Name(), "%s")

	g.Printf("\n\n")

	return nil
}

// Name returns struct type name.
func (st *BuildStruct) Name() string {
	return st.typeInfo.Name()
}

// Kind returns kind from struct.
func (st *BuildStruct) Kind() string {
	for _, field := range st.Fields {
		if field.Tag.Kind != "" {
			return field.Tag.Kind
		}
	}
	return st.Name()
}

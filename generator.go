package qbg

import (
	"bytes"
	"reflect"
	"strings"
	"text/template"
	"unicode"

	"github.com/favclip/genbase"
)

// BuildSource represents source code of assembling..
type BuildSource struct {
	g         *genbase.Generator
	pkg       *genbase.PackageInfo
	typeInfos genbase.TypeInfos

	InlineInterfaces    bool
	UseDatastoreWrapper bool
	Structs             []*BuildStruct
}

// BuildStruct represents struct of assembling..
type BuildStruct struct {
	parent   *BuildSource
	typeInfo *genbase.TypeInfo

	Private bool
	Fields  []*BuildField
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
// deprecated. use *BuildSource#Parse instead.
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

// Parse construct *BuildSource from package & type information.
func (b *BuildSource) Parse(pkg *genbase.PackageInfo, typeInfos genbase.TypeInfos) error {
	if b.g == nil {
		b.g = genbase.NewGenerator(pkg)
	}
	b.pkg = pkg
	b.typeInfos = typeInfos

	if b.UseDatastoreWrapper {
		b.g.AddImport("go.mercari.io/datastore", "")
	} else {
		b.g.AddImport("google.golang.org/appengine/datastore", "")
	}
	if !b.InlineInterfaces {
		b.g.AddImport("github.com/favclip/qbg/qbgutils", "")
	}

	for _, typeInfo := range typeInfos {
		err := b.parseStruct(typeInfo)
		if err != nil {
			return err
		}
	}

	return nil
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
			} else if key == "goon" || key == "boom" {
				tagText := structTag.Get("goon")
				if tagText == "" {
					tagText = structTag.Get("boom")
				}
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

	if b.InlineInterfaces {
		tmpl := template.New("plugin")
		tmpl, err := tmpl.Parse(pluginTemplate)
		if err != nil {
			return nil, err
		}

		var dsKeyType string
		if b.UseDatastoreWrapper {
			dsKeyType = "datastore.Key"
		} else {
			dsKeyType = "*datastore.Key"
		}

		buf := bytes.NewBufferString("")
		err = tmpl.Execute(buf, map[string]string{
			"DSKeyType": dsKeyType,
		})
		if err != nil {
			return nil, err
		}

		b.g.Printf(buf.String())
	}

	for _, st := range b.Structs {
		err := st.emit(b.g)
		if err != nil {
			return nil, err
		}
	}

	return b.g.Format()
}

func (st *BuildStruct) emit(g *genbase.Generator) error {
	tmpl := template.New("struct")
	tmpl, err := tmpl.Parse(structTemplate)
	if err != nil {
		return err
	}

	var newWord string
	if st.Private {
		newWord = "new"
	} else {
		newWord = "New"
	}
	var pluginType string
	if st.parent.InlineInterfaces {
		pluginType = "Plugin"
	} else {
		pluginType = "qbgutils.Plugin"
	}
	var pluggerType string
	if st.parent.InlineInterfaces {
		pluggerType = "Plugger"
	} else {
		pluggerType = "qbgutils.Plugger"
	}
	var fields []*BuildField
	for _, f := range st.Fields {
		if f.Tag.Ignore {
			continue
		}
		fields = append(fields, f)
	}
	var dsKeyType string
	var dsQueryType string
	var dsNewQuery string
	if st.parent.UseDatastoreWrapper {
		dsKeyType = "datastore.Key"
		dsQueryType = "datastore.Query"
		dsNewQuery = "client.NewQuery"
	} else {
		dsKeyType = "*datastore.Key"
		dsQueryType = "*datastore.Query"
		dsNewQuery = "datastore.NewQuery"
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, map[string]interface{}{
		"UserDatastoreWrapper": st.parent.UseDatastoreWrapper,
		"Name":                 st.Name(),
		"SimpleName":           st.SimpleName(),
		"Kind":                 st.Kind(),
		"NewWord":              newWord,
		"DSKeyType":            dsKeyType,
		"DSQueryType":          dsQueryType,
		"DSNewQuery":           dsNewQuery,
		"PluginType":           pluginType,
		"PluggerType":          pluggerType,
		"Fields":               fields,
	})
	if err != nil {
		return err
	}

	g.Printf(buf.String())

	return nil
}

// Name returns struct type name.
func (st *BuildStruct) Name() string {
	// FooBar -> fooBar
	// IIS -> iis
	// AEData -> aeData

	name := st.SimpleName()

	if !st.Private {
		return name
	}
	if len(name) <= 1 {
		return strings.ToLower(name)
	}
	if name == strings.ToUpper(name) {
		return strings.ToLower(name)
	}

	runeNames := []rune(name)
	if unicode.IsLower(runeNames[0]) {
		return name
	} else if unicode.IsLower(runeNames[1]) {
		return string(unicode.ToLower(runeNames[0])) + string(runeNames[1:])
	}

	var idx int
	for idx = 0; idx < len(runeNames); idx++ {
		r := runeNames[idx]
		if unicode.IsLower(r) {
			break
		}
	}

	return strings.ToLower(string(runeNames[0:idx-1])) + string(runeNames[idx-1:])
}

// Name returns struct type name.
func (st *BuildStruct) SimpleName() string {
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

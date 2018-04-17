package internal

import (
	"log"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/introspection"
)

const (
	gqlSCALAR       = "SCALAR"
	gqlINTERFACE    = "INTERFACE"
	gqlINPUT_OBJECT = "INPUT_OBJECT"
	gqlUNION        = "UNION"
	gqlLIST         = "LIST"
	gqlOBJECT       = "OBJECT"
	gqlENUM         = "ENUM"
	gqlDIRECTIVE    = "DIRECTIVE"
	gqlID           = "ID"
	gqlQuery        = "Query"
	gqlMutation     = "Mutation"
)

var KnownGQLTypes = map[string]bool{
	"__Directive":         true,
	"__DirectiveLocation": true,
	"__EnumValue":         true,
	"__Field":             true,
	"__InputValue":        true,
	"__Schema":            true,
	"__Type":              true,
	"__TypeKind":          true,
	"LIST":                true,
	"String":              true,
	"Float":               true,
	"ID":                  true,
	"Int":                 true,
	"Boolean":             true,
	"Time":                true,
}

var KnownGoTypes = map[string]bool{
	"string":    true,
	"bool":      true,
	"float32":   true,
	"float64":   true,
	"int":       true,
	"int32":     true,
	"int64":     true,
	"time.Time": true,
}

type CodeGen struct {
	rawSchema string

	schema *graphql.Schema
}

type TypeDef struct {
	Name        string
	Description string
	GQLType     string
	Template    string

	Fields     []*FieldDef
	Efields    []*FieldDef // Extra Fields - Model Userdefined Fields ex: Todos, Posts on User type
	Interfaces map[string]string

	IsQuery     bool
	IsMutation  bool
	IsInterface bool
	IsObject    bool
	IsInput     bool
	IsScalar    bool
	IsModel     bool

	gqlType *introspection.Type
}

type Typ struct {
	GoType  string
	GQLType string

	Type   *Typ
	Values []string

	IsEntry     bool
	IsNullable  bool
	IsInterface bool
	IsInput     bool

	gqlType *introspection.Type
}

type FieldDef struct {
	Name        string
	Parent      string
	Description string

	IsDeprecated      bool
	DeprecationReason string

	Type *Typ
	Args []*FieldDef

	IsInterface   bool
	IsInput       bool
	IsUserDefined bool

	gqlField *introspection.Field
}

func NewType(t *introspection.Type) *TypeDef {

	tp := &TypeDef{
		Name:        pts(t.Name()),
		Description: pts(t.Description()),
		Fields:      []*FieldDef{},
		Efields:     []*FieldDef{},
		Interfaces:  map[string]string{},
		gqlType:     t,
	}

	/**
	 * union & input object types do not have fields
	 * so we ignore it to avoid nil pointer dereference error
	 * for input object type we create fields from InputFields instead
	 */
	if t.Kind() != gqlUNION && t.Kind() != gqlINPUT_OBJECT {
		// Get Depreceated fields also
		fd := &struct{ IncludeDeprecated bool }{true}

		if t.Fields(fd) != nil {
			for _, fld := range *t.Fields(fd) {
				f := NewField(fld)
				f.Parent = tp.Name
				f.IsInterface = t.Kind() == gqlINTERFACE
				tp.Fields = append(tp.Fields, f)

				// if tp.Name == "User" {
				// 	log.Printf("Field %v", t)
				// }
			}
		}

		if t.Kind() == gqlOBJECT {
			interfaces := *t.Interfaces()
			for _, i := range interfaces {
				name := pts(i.Name())
				tp.Interfaces[name] = name
			}
		}
	} else if t.Kind() == gqlINPUT_OBJECT {
		for _, input := range *t.InputFields() {
			f := newField(input.Name(), input.Description(), input.Type())
			f.IsInput = t.Kind() == gqlINPUT_OBJECT
			f.Parse()
			tp.Fields = append(tp.Fields, f)
		}
	}

	return tp
}

func newField(name string, desc *string, typ *introspection.Type) *FieldDef {
	return &FieldDef{
		Name:        fieldName(name),
		Description: pts(desc),
		Type: &Typ{
			IsNullable: true,
			gqlType:    typ,
		},
	}
}

func NewField(t *introspection.Field) *FieldDef {

	fld := newField(t.Name(), t.Description(), t.Type())
	fld.Parse()

	// parse arguments (i.e., interface function)
	for _, arg := range t.Args() {
		argFld := newField(arg.Name(), arg.Description(), arg.Type())
		argFld.Parse()
		fld.Args = append(fld.Args, argFld)
	}

	return fld
}

func (f *FieldDef) Parse() {

	tp := f.Type.gqlType
	td := f.Type

FindGoType:
	td.gqlType = tp
	if tp.Kind() == "NON_NULL" {
		td.IsNullable = false
		tp = tp.OfType()
	}

	if tp.Kind() == "LIST" {
		td.GoType = "[]"
		td.GQLType = "[]"

		td.Type = &Typ{
			IsNullable: true,
		}

		tp = tp.OfType()
		td = td.Type
		goto FindGoType
	}

	switch *tp.Name() {
	case "String":
		td.GoType = "string"
		td.GQLType = "string"
	case "Int":
		td.GoType = "int32"
		td.GQLType = "int32"
	case "Float":
		td.GoType = "float64"
		td.GQLType = "float64"
	case "ID":
		// TODO - shouldn't we use graphql.ID type for `ID` fields
		// because it may not work for query and mutation calls?
		td.GoType = "string"
		td.GQLType = "graphql.ID"
	case "Boolean":
		td.GoType = "bool"
		td.GQLType = "bool"
	case "Time":
		td.GoType = "time.Time"
		td.GQLType = "graphql.Time"
	default:
		if tp.Kind() == gqlENUM {
			td.GoType = "string"
			td.GQLType = "string"
		} else {
			td.GoType = pts(tp.Name())
			td.GQLType = pts(tp.Name()) + "Resolver"
			//td.IsUserDefined = true
		}
	}
	return
}

func NewCodeGen(schema string) *CodeGen {
	return &CodeGen{schema, nil}
}

func (g *CodeGen) Parse() error {
	schema, err := graphql.ParseSchema(g.rawSchema, nil)
	g.schema = schema
	return err
}

func (g *CodeGen) Generate(args *ArgType) error {
	// Parse the sting schema to schema object
	g.Parse()

	// Types for Generation
	types := []*TypeDef{}

	// Types that implements Node - Useful for extra work / model creation etc
	models := map[string]*TypeDef{}

	// Mutation Payloads
	payloads := map[string]*TypeDef{}

	for _, typ := range g.schema.Inspect().Types() {
		if KnownGQLTypes[*typ.Name()] {
			continue
		}

		//log.Printf("Generating Go code for %s %s", typ.Kind(), pts(typ.Name()))
		switch typ.Kind() {
		case gqlOBJECT:
			gtp := NewType(typ)
			typName := pts(typ.Name())

			// Identify Query & Mutation definitions
			if typName == gqlQuery || typName == gqlMutation {
				gtp.IsQuery = typName == gqlQuery
				gtp.IsMutation = typName == gqlMutation
			}

			// Save payloads to string array for better handling
			if strings.HasSuffix(typName, "Payload") {
				//log.Printf("Graphql type %s", typName)
				payloads[gtp.Name] = gtp
			} else {
				gtp.IsObject = true
				types = append(types, gtp)
			}
		case gqlSCALAR:
			gtp := NewType(typ)
			gtp.IsScalar = true
			//types = append(types, gtp)
		case gqlINTERFACE:
			gtp := NewType(typ)
			gtp.IsInterface = true
			types = append(types, gtp)
		case gqlENUM:
		case gqlUNION:
		case gqlINPUT_OBJECT:
			gtp := NewType(typ)
			gtp.IsInput = true
			types = append(types, gtp)
		default:
			log.Printf("Unknown graphql type ", *typ.Name(), ":", typ.Kind())
		}
	}

	for _, t := range types {
		// Set models
		if t.IsObject && !t.IsMutation && !t.IsQuery {
			for _, i := range t.Interfaces {
				// Get Node implemented types - We can use this info create MODELS etc
				if i == "Node" {
					// Set model is true based on [TYPE] implements [NODE]
					t.IsModel = true

					// Save them to a map for easy access
					models[t.Name] = t
				}
			}
		}
	}

	for _, t := range types {
		// For types
		if t.IsObject && !t.IsMutation && !t.IsQuery {

			// Generate Types
			if err := g.generateType(args, t, models); err != nil {
				return err
			}
		}

		// For Queries & Mutation
		if t.IsQuery || t.IsMutation {

			// Generate Query/Mutation
			if err := g.generateQuery(args, t, models, false); err != nil {
				return err
			}
		}

		// For InputObjects
		if t.IsInput {

			// Generate Query/Mutation
			if err := g.generateQuery(args, t, models, false); err != nil {
				return err
			}
		}
	}

	for _, p := range payloads {
		// Generate Mutation Payload
		if err := g.generateQuery(args, p, models, true); err != nil {
			return err
		}
	}

	return nil
}

func (g *CodeGen) generateQuery(args *ArgType, tp *TypeDef, models map[string]*TypeDef, payload bool) error {

	if tp.IsQuery {
		log.Printf("Generating Go code for QUERY [%s]", tp.Name)

		tp.Template = "QUERY"
		err := args.ExecuteTemplate(GraphQLQueryTemplate, "iQuery", gqlQuery, tp)
		if err != nil {
			return err
		}
	}

	if tp.IsMutation {
		log.Printf("Generating Go code for MUTATION [%s]", tp.Name)

		tp.Template = "MUTATION"
		err := args.ExecuteTemplate(GraphQLQueryTemplate, "iMutation", gqlMutation, tp)
		if err != nil {
			return err
		}
	}

	if tp.IsMutation {
		log.Printf("Generating Go code for MUTATION ARGS [%s]", tp.Name)

		tp.Template = "ARGS"
		err := args.ExecuteTemplate(GraphQLQueryTemplate, "request", gqlMutation, tp)
		if err != nil {
			return err
		}
	}

	// Mutation inputs
	if tp.IsInput {
		log.Printf("Generating Go code for INPUT [%s]", tp.Name)

		tp.Template = "INPUT"
		err := args.ExecuteTemplate(GraphQLQueryTemplate, "request", gqlMutation, tp)
		if err != nil {
			return err
		}
	}

	// Mutation Payloads
	if payload {
		log.Printf("Generating Go code for PAYLOAD [%s]", tp.Name)

		tp.Template = "PAYLOAD"
		err := args.ExecuteTemplate(GraphQLQueryTemplate, "response", gqlMutation, tp)
		if err != nil {
			return err
		}
	}

	return nil
}

// Generate Type
func (g *CodeGen) generateType(args *ArgType, tp *TypeDef, models map[string]*TypeDef) error {

	log.Printf("Generating Go code for TYPE [%s]", tp.Name)

	// Check Model Explicitly
	if tp.IsModel {
		for i := len(tp.Fields) - 1; i >= 0; i-- {
			// get field by new length of fields
			f := tp.Fields[i]

			// Check if the field is GoType or User defined for Models
			if _, ok := KnownGoTypes[f.Type.GoType]; !ok {
				// Mark this field as userDefined
				f.IsUserDefined = true

				// Remove this field from - Fields List
				tp.Fields = append(tp.Fields[:i], tp.Fields[i+1:]...)

				// Add this field to - Extra Fields List
				tp.Efields = append(tp.Efields, f)
			}
		}

		// // generate TYPE_EXTRA - Model's Additional Fields
		// tp.Template = "EXTRA"
		// tplName := tp.Name + "_extra"

		// err := args.ExecuteTemplate(GraphQLTypeTemplate, tplName, gqlOBJECT, tp)
		// if err != nil {
		// 	return err
		// }
	}

	// generate TYPE
	tp.Template = "TYPE"
	templateName := tp.Name

	templateName = strings.TrimSuffix(templateName, "Edge")
	templateName = strings.TrimSuffix(templateName, "Connection")

	err := args.ExecuteTemplate(GraphQLTypeTemplate, templateName, gqlOBJECT, tp)
	if err != nil {
		return err
	}

	return nil
}

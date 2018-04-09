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
	"string":  true,
	"bool":    true,
	"float32": true,
	"float64": true,
	"int":     true,
	"int32":   true,
	"int64":   true,
}

type CodeGen struct {
	rawSchema    string
	mutationName string
	queryName    string

	schema *graphql.Schema
}

type TypeDef struct {
	Name        string
	Description string
	GQLType     string
	Fields      []*FieldDef
	Interfaces  map[string]string

	IsQuery     bool
	IsMutation  bool
	IsScalar    bool
	IsInterface bool
	IsInput     bool
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
		td.GoType = "float32"
		td.GQLType = "float32"
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
	return &CodeGen{schema, "Mutation", "Query", nil}
}

func (g *CodeGen) Parse() error {
	schema, err := graphql.ParseSchema(g.rawSchema, nil)
	g.schema = schema
	return err
}

func (g *CodeGen) Generate(args *ArgType) error {
	// Parse the sting schema to schema object
	g.Parse()

	inspect := g.schema.Inspect()

	if inspect.MutationType() != nil {
		g.mutationName = pts(inspect.MutationType().Name())
	}

	if inspect.QueryType() != nil {
		g.queryName = pts(inspect.QueryType().Name())
	}

	// Types for Generation
	types := []*TypeDef{}

	// Types that implements Node - Useful for extra work / model creation etc
	models := map[string]*TypeDef{}

	for _, typ := range inspect.Types() {
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

			types = append(types, gtp)
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
			//types = append(types, gtp)
		default:
			log.Printf("Unknown graphql type ", *typ.Name(), ":", typ.Kind())
		}
	}

	for _, t := range types {
		if !t.IsInterface && !t.IsQuery && !t.IsMutation {
			for _, i := range t.Interfaces {
				// Get Node implemented types - We can use this info create MODELS etc
				if i == "Node" {
					// Set model is true based on [TYPE] implement [NODE]
					t.IsModel = true

					// Save them to a map for easy access
					models[t.Name] = t
				}
			}

			// Generate Types
			if err := g.generateType(args, t, models); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *CodeGen) generateType(args *ArgType, tp *TypeDef, models map[string]*TypeDef) error {

	log.Printf("Generating Go code for %s %s", gqlOBJECT, tp.Name)

	templateType := GraphQLTypeTemplate
	templateName := tp.Name

	templateName = strings.TrimSuffix(templateName, "Edge")
	templateName = strings.TrimSuffix(templateName, "Connection")

	// Check Model Explicitly
	if tp.IsModel {
		for _, f := range tp.Fields {
			// Check if the field is GoType or User defined for Models
			if ok := KnownGoTypes[f.Type.GoType]; !ok {
				f.IsUserDefined = true
			}

			// Special case for Graphql Time and ID
			if f.Type.GoType == "time.Time" || f.Type.GoType == "graphql.Time" || f.Type.GoType == "graphql.ID" {
				f.IsUserDefined = false
			}
		}
	}

	// generate type template
	err := args.ExecuteTemplate(templateType, templateName, gqlOBJECT, tp)
	if err != nil {
		return err
	}

	return nil
	// name := *tp.Name()
	// templateType := GraphQLTypeTemplate
	// templateName := name

	// if name == g.queryName {
	// 	templateType = GraphQLQueryTemplate
	// }

	// if name == g.mutationName {
	// 	templateType = GraphQLMutationTemplate
	// }

	// if tp.Kind() == "INPUT_OBJECT" {
	// 	templateType = GraphQLMutationTemplate
	// }

	// if tp.Kind() == "OBJECT" && strings.HasSuffix(templateName, "Payload") {
	// 	templateType = GraphQLMutationTemplate
	// }

	// templateName = strings.TrimSuffix(name, "Edge")
	// templateName = strings.TrimSuffix(templateName, "Connection")

	// // if verbose
	// if args.Verbose {
	// 	log.Printf("GenerateType [%s] Template Name [%s] - QUERY [%t]", name, templateName, (name == g.queryName))
	// }

	// // Move this to a util func (g *CodeGen)
	// var ifields []*introspection.Field
	// if tp.Fields(&struct{ IncludeDeprecated bool }{true}) != nil {
	// 	ifields = *tp.Fields(&struct{ IncludeDeprecated bool }{true})
	// }

	// fields := make([]GqlField, len(ifields))
	// methods := make([]GqlMethod, len(ifields))

	// for i, fp := range ifields {
	// 	fieldCode, methodCode := g.generateField(args, fp, tp)
	// 	fields[i] = fieldCode
	// 	methods[i] = methodCode
	// }

	// var inputFields []GqlField
	// if tp.InputFields() != nil {
	// 	for _, ip := range *tp.InputFields() {
	// 		inputField := g.generateInputValue(args, ip, tp)
	// 		inputFields = append(inputFields, inputField)
	// 	}
	// }

	// possibleTypes := []string{}
	// if tp.PossibleTypes() != nil {
	// 	for _, tp := range *tp.PossibleTypes() {
	// 		possibleTypes = append(possibleTypes, *tp.Name())
	// 	}
	// }

	// enumValues := []string{}
	// if tp.EnumValues(&struct{ IncludeDeprecated bool }{true}) != nil {
	// 	for _, value := range *tp.EnumValues(&struct{ IncludeDeprecated bool }{true}) {
	// 		enumValues = append(enumValues, value.Name())
	// 	}
	// }

	// typeTpl := map[string]interface{}{
	// 	"Kind":            tp.Kind(),
	// 	"PossibleTypes":   possibleTypes,
	// 	"EnumValues":      enumValues,
	// 	"TypeName":        name,
	// 	"TypeDescription": args.removeLineBreaks(g.returnString(tp.Description())),
	// 	"Fields":          fields,
	// 	"InputFields":     inputFields,
	// 	"Methods":         methods,
	// 	"IsEntry":         g.isEntryPoint(name),
	// 	"IsScalar":        tp.Kind() == "SCALAR",
	// 	"IsInterface":     tp.Kind() == "INTERFACE",
	// 	"IsInput":         tp.Kind() == "INPUT_OBJECT",
	// 	"IsResolver":      tp.Kind() == "RESOLVER",
	// }

	// //log.Printf("TYPEKIND [%s]", tp.Kind())

	// if templateType == GraphQLQueryTemplate {
	// 	for _, m := range methods {
	// 		templateName = args.unCapitalise(m.ReturnType)
	// 		templateName = strings.TrimPrefix(templateName, "*")

	// 		templateName = strings.TrimSuffix(templateName, "Resolver")
	// 		templateName = strings.TrimSuffix(templateName, "Connection")

	// 		// override methods to single method in loop
	// 		typeTpl["Methods"] = append([]GqlMethod{}, m)
	// 		m.IsQuery = true

	// 		// if verbose
	// 		if args.Verbose {
	// 			log.Printf("GenerateQuery [%s] Template Name [%s] MethodName [%s]", name, templateName, m.Name)
	// 		}

	// 		// generate query template
	// 		err := args.ExecuteTemplate(templateType, templateName, tp.Kind(), typeTpl)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// } else if templateType == GraphQLMutationTemplate {
	// 	if templateName == "Mutation" {
	// 		for _, m := range methods {
	// 			templateName = args.unCapitalise(m.Name)
	// 			m.NReturnType = strings.TrimSuffix(m.NReturnType, "Resolver")
	// 			m.IsMutation = true

	// 			// override methods to single method in loop
	// 			typeTpl["Methods"] = append([]GqlMethod{}, m)

	// 			// if verbose
	// 			if args.Verbose {
	// 				log.Printf("GenerateMutation [%s] Template Name [%s] MethodName [%s]", name, templateName, m.Name)
	// 			}

	// 			// generate type template
	// 			err := args.ExecuteTemplate(templateType, templateName, tp.Kind(), typeTpl)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	// 	} else {
	// 		// reset the old methods
	// 		typeTpl["Methods"] = methods

	// 		templateName = strings.TrimSuffix(templateName, "Input")
	// 		templateName = strings.TrimSuffix(templateName, "Payload")

	// 		// if verbose
	// 		if args.Verbose {
	// 			log.Printf("GenerateMutation [%s] Template Name [%s] MethodName [%s]", name, templateName, "")
	// 		}

	// 		// generate type template
	// 		err := args.ExecuteTemplate(templateType, templateName, tp.Kind(), typeTpl)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// } else {
	// 	// reset the old methods
	// 	typeTpl["Methods"] = methods

	// 	// generate type template
	// 	err := args.ExecuteTemplate(templateType, templateName, tp.Kind(), typeTpl)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
}

func (g *CodeGen) getTypeName(tp *introspection.Type, input bool) (typ string) {
	// check:
	// 	if tp.Kind() == "NON_NULL" {
	// 		tp = tp.OfType()
	// 	} else {
	// 		typ = typ + "*"
	// 	}

	// 	if tp.Kind() == "LIST" {
	// 		tp = tp.OfType()
	// 		typ = typ + "[]"
	// 		goto check
	// 	}

	// 	name := tp.Name()
	// 	if val, ok := internalTypeConfig[*name]; ok {
	// 		return typ + val.goType
	// 	}

	// 	if tp.Kind() == "ENUM" {
	// 		typ = typ + *name
	// 	} else if tp.Kind() != "INPUT_OBJECT" {
	// 		if len(typ) > 0 {
	// 			if typ[len(typ)-1] != '*' {
	// 				typ = typ + "*"
	// 			}
	// 		} else if tp.Kind() != "SCALAR" {
	// 			typ = "*"
	// 		}
	// 		typ = typ + *name + "Resolver"
	// 	} else {
	// 		typ = typ + *name
	// 	}

	// 	if typ[0] != '*' && tp.Kind() == "INPUT_OBJECT" {
	// 		typ = "*" + typ
	// 	}

	return
}

func (g *CodeGen) isEntryPoint(a string) bool {
	return a == g.mutationName || a == g.queryName
}

// Helper functions

func fieldName(name string) string {
	//name = upperFirst(name)
	if name == "Id" || name == "id" || name == "ID" {
		name = "ID"
	}

	return name
}

func isNullable(fp *introspection.Field) bool {
	if fp.Type().Kind() == "NON_NULL" {
		return false
	}

	return true
}

func pts(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func gqlReturnType(t *Typ, mode, fieldName, pkg string) string {
	var r string

	if mode == "struct" || mode == "argStruct" {
		r = t.GoType

		ok := KnownGoTypes[t.GoType]
		if !ok {
			r = t.GQLType
		}
	} else {
		r = t.GQLType
	}

	if mode == "struct" {
		ok := KnownGoTypes[t.GoType]
		if t.IsNullable && t.GQLType != "[]" && !ok {
			r = "*" + r
		}
	} else {
		if t.IsNullable {
			r = "*" + r
		}
	}

	// Special case for pageInfo - ugly hack
	if (mode == "struct" || mode == "resolver") && fieldName == "pageInfo" {
		r = "*" + r
	}

	if t.Type == nil {
		return r
	}

	r += gqlReturnType(t.Type, mode, fieldName, pkg)

	// Special case for edges - ugly hack
	if mode == "struct" && fieldName == "edges" {
		r = "*" + r
	}

	return r
}

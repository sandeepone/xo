package internal

import (
	"log"
	"strings"

	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/introspection"
)

type typeConfig struct {
	ignore bool
	goType string
}

var (
	internalTypeConfig = map[string]typeConfig{
		"SCALAR": typeConfig{
			true,
			"",
		},
		"Boolean": typeConfig{
			true,
			"bool",
		},
		"Float": typeConfig{
			true,
			"float64",
		},
		"Int": typeConfig{
			true,
			"int32",
		},
		"ID": typeConfig{
			true,
			"graphql.ID",
		},
		"Time": typeConfig{
			true,
			"graphql.Time",
		},
		"String": typeConfig{
			true,
			"string",
		},
	}
)

type CodeGen struct {
	graphSchema  string
	mutationName string
	queryName    string
}

type GqlField struct {
	Name        string
	Description string
	FieldType   string
	TypeKind    string
	TypeName    string
	IsEntry     bool
	IsNullable  bool
}

type FieldArgument struct {
	Name string
	Type string
}

type GqlMethod struct {
	Name        string
	Description string
	TypeKind    string
	TypeName    string
	Arguments   []FieldArgument
	Return      string
	ReturnType  string
	NReturnType string
	IsEntry     bool
	IsNullable  bool
}

func NewCodeGen(graphSchema string) *CodeGen {
	return &CodeGen{graphSchema, "", ""}
}

func (g *CodeGen) Generate(args *ArgType) error {
	graphSchema := g.graphSchema

	sch, err := graphql.ParseSchema(graphSchema, nil)
	if err != nil {
		return err
	}

	ins := sch.Inspect()

	if ins.MutationType() != nil {
		g.mutationName = g.returnString(ins.MutationType().Name())
	}

	if ins.QueryType() != nil {
		g.queryName = g.returnString(ins.QueryType().Name())
	}

	var entryPoint = false

	for _, qlType := range ins.Types() {
		name := *qlType.Name()
		if strings.HasPrefix(name, "_") {
			continue
		}

		if internalTypeConfig[name].ignore {
			continue
		}

		if g.isEntryPoint(name) {
			entryPoint = true
		}

		//log.Printf("Generating Go code for %s %s", qlType.Kind(), name)

		err = g.generateType(args, qlType)
		if err != nil {
			return err
		}
	}

	// Generate entry point
	if entryPoint {
		err := g.generateEntryPoint(args)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *CodeGen) returnString(strPtr *string) string {
	if strPtr != nil {
		return *strPtr
	}
	return ""
}

func (g *CodeGen) generateEntryPoint(args *ArgType) error {
	typeTpl := map[string]interface{}{
		"Kind":            "RESOLVER",
		"TypeName":        "Resolver",
		"TypeDescription": "Resolver is the main resolver for all queries",
	}

	// generate query type template
	err := args.ExecuteTemplate(GraphQLTypeTemplate, "default", "RESOLVER", typeTpl)
	if err != nil {
		return err
	}

	return nil
}

func (g *CodeGen) generateType(args *ArgType, tp *introspection.Type) error {
	name := *tp.Name()
	templateType := GraphQLTypeTemplate
	templateName := name

	if name == g.queryName {
		templateType = GraphQLQueryTemplate
	}

	if name == g.mutationName {
		templateType = GraphQLMutationTemplate
	}

	templateName = strings.TrimSuffix(name, "Edge")
	templateName = strings.TrimSuffix(templateName, "Connection")

	// if verbose
	if args.Verbose {
		log.Printf("GenerateType [%s] Template Name [%s] - QUERY [%t]", name, templateName, (name == g.queryName))
	}

	// Move this to a util func (g *CodeGen)
	var ifields []*introspection.Field
	if tp.Fields(&struct{ IncludeDeprecated bool }{true}) != nil {
		ifields = *tp.Fields(&struct{ IncludeDeprecated bool }{true})
	}

	fields := make([]GqlField, len(ifields))
	methods := make([]GqlMethod, len(ifields))

	for i, fp := range ifields {
		fieldCode, methodCode := g.generateField(args, fp, tp)
		fields[i] = fieldCode
		methods[i] = methodCode
	}

	var inputFields []GqlField
	if tp.InputFields() != nil {
		for _, ip := range *tp.InputFields() {
			inputField := g.generateInputValue(args, ip, tp)
			inputFields = append(inputFields, inputField)
		}
	}

	possibleTypes := []string{}
	if tp.PossibleTypes() != nil {
		for _, tp := range *tp.PossibleTypes() {
			possibleTypes = append(possibleTypes, *tp.Name())
		}
	}

	enumValues := []string{}
	if tp.EnumValues(&struct{ IncludeDeprecated bool }{true}) != nil {
		for _, value := range *tp.EnumValues(&struct{ IncludeDeprecated bool }{true}) {
			enumValues = append(enumValues, value.Name())
		}
	}

	typeTpl := map[string]interface{}{
		"Kind":            tp.Kind(),
		"PossibleTypes":   possibleTypes,
		"EnumValues":      enumValues,
		"TypeName":        name,
		"TypeDescription": args.removeLineBreaks(g.returnString(tp.Description())),
		"Fields":          fields,
		"InputFields":     inputFields,
		"Methods":         methods,
		"IsEntry":         g.isEntryPoint(name),
	}

	if templateType == GraphQLQueryTemplate {
		for _, m := range methods {
			templateName = args.unCapitalise(m.ReturnType)
			templateName = strings.TrimPrefix(templateName, "*")

			templateName = strings.TrimSuffix(templateName, "Resolver")
			templateName = strings.TrimSuffix(templateName, "Connection")

			// override methods to single method in loop
			typeTpl["Methods"] = append([]GqlMethod{}, m)

			// if verbose
			if args.Verbose {
				log.Printf("GenerateQuery [%s] Template Name [%s] MethodName [%s]", name, templateName, m.Name)
			}

			// generate query template
			err := args.ExecuteTemplate(templateType, templateName, tp.Kind(), typeTpl)
			if err != nil {
				return err
			}
		}
	} else if templateType == GraphQLMutationTemplate {
		// if verbose
		if args.Verbose {
			log.Printf("GenerateMutation [%s] Template Name [%s] MethodName [%s]", name, templateName, "")
		}
	} else {
		// reset the old methods
		typeTpl["Methods"] = methods

		// generate type template
		err := args.ExecuteTemplate(templateType, templateName, tp.Kind(), typeTpl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *CodeGen) generateInputValue(args *ArgType, ip *introspection.InputValue, tp *introspection.Type) GqlField {
	name := ip.Name()
	fieldTypeName := g.getTypeName(ip.Type(), false)

	return GqlField{
		Name:        name,
		Description: args.removeLineBreaks(g.returnString(ip.Description())),
		TypeKind:    tp.Kind(),
		TypeName:    name,
		FieldType:   fieldTypeName,
		IsEntry:     g.isEntryPoint(name),
	}

}

func (g *CodeGen) generateField(args *ArgType, fp *introspection.Field, tp *introspection.Type) (GqlField, GqlMethod) {
	name := fp.Name()
	typeName := *tp.Name()

	fieldTypeName := g.getTypeName(fp.Type(), false)
	fieldArguments := make([]FieldArgument, 0, len(fp.Args()))

	for _, field := range fp.Args() {
		fieldArguments = append(fieldArguments, FieldArgument{
			Name: field.Name(),
			Type: g.getTypeName(field.Type(), true),
		})
	}

	gqlField := GqlField{
		Name:        name,
		Description: args.removeLineBreaks(g.returnString(fp.Description())),
		FieldType:   fieldTypeName,
		TypeKind:    tp.Kind(),
		TypeName:    typeName,
		IsEntry:     g.isEntryPoint(name),
		IsNullable:  g.isNullable(fp),
	}

	gqlMethod := GqlMethod{
		Name:        name,
		Description: args.removeLineBreaks(g.returnString(fp.Description())),
		TypeKind:    tp.Kind(),
		TypeName:    typeName,
		Arguments:   fieldArguments,
		Return:      name,
		ReturnType:  fieldTypeName,
		NReturnType: strings.Replace(fieldTypeName, "*", "", 1),
		IsEntry:     g.isEntryPoint(typeName),
		IsNullable:  g.isNullable(fp),
	}

	return gqlField, gqlMethod
}

func (g *CodeGen) getTypeName(tp *introspection.Type, input bool) (typ string) {
check:
	if tp.Kind() == "NON_NULL" {
		tp = tp.OfType()
	} else {
		typ = typ + "*"
	}

	if tp.Kind() == "LIST" {
		tp = tp.OfType()
		typ = typ + "[]"
		goto check
	}

	name := tp.Name()
	if val, ok := internalTypeConfig[*name]; ok {
		return typ + val.goType
	}

	if tp.Kind() == "ENUM" {
		typ = typ + *name
	} else if tp.Kind() != "INPUT_OBJECT" {
		if len(typ) > 0 {
			if typ[len(typ)-1] != '*' {
				typ = typ + "*"
			}
		} else if tp.Kind() != "SCALAR" {
			typ = "*"
		}
		typ = typ + *name + "Resolver"
	} else {
		typ = typ + *name
	}

	if typ[0] != '*' && tp.Kind() == "INPUT_OBJECT" {
		typ = "*" + typ
	}

	return
}

func (g *CodeGen) isNullable(fp *introspection.Field) bool {
	if fp.Type().Kind() == "NON_NULL" {
		return false
	}

	return true
}

func (g *CodeGen) isEntryPoint(a string) bool {
	return a == g.mutationName || a == g.queryName
}

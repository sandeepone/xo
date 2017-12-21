package internal

import (
	"bytes"
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

		log.Printf("Generating Go code for %s %s", qlType.Kind(), name)

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

	templateName := strings.TrimSuffix(name, "Edge")
	templateName = strings.TrimSuffix(templateName, "Connection")

	// if verbose
	if args.Verbose {
		log.Printf("GenerateType [%s] Template Name [%s]", name, templateName)
	}

	// Move this to a util func (g *CodeGen)
	var ifields []*introspection.Field
	if tp.Fields(&struct{ IncludeDeprecated bool }{true}) != nil {
		ifields = *tp.Fields(&struct{ IncludeDeprecated bool }{true})
	}

	fields := make([]string, len(ifields))
	methods := make([]string, len(ifields))

	for i, fp := range ifields {
		fieldCode, methodCode, err := g.generateField(args, fp, tp)
		if err != nil {
			return err
		}
		fields[i] = fieldCode
		methods[i] = methodCode
	}

	var inputFields []string
	if tp.InputFields() != nil {
		for _, ip := range *tp.InputFields() {
			inputField, err := g.generateInputValue(args, ip, tp)
			if err != nil {
				return err
			}
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
	}

	// generate query type template
	err := args.ExecuteTemplate(GraphQLTypeTemplate, templateName, tp.Kind(), typeTpl)
	if err != nil {
		return err
	}

	return nil
}

func (g *CodeGen) generateInputValue(args *ArgType, ip *introspection.InputValue, tp *introspection.Type) (string, error) {
	name := ip.Name()

	fieldTypeName := g.getTypeName(ip.Type(), false)

	fieldTpl := map[string]interface{}{
		"TypeKind":         tp.Kind(),
		"FieldName":        name,
		"FieldDescription": args.removeLineBreaks(g.returnString(ip.Description())),
		"FieldType":        fieldTypeName,
	}

	// generate field template
	fieldCode, err := args.ExecuteTemplateBuffer(GraphQLFieldTemplate, name, fieldTypeName, fieldTpl)
	if err != nil {
		return "", err
	}

	return string(fieldCode.Bytes()), nil
}

func (g *CodeGen) generateField(args *ArgType, fp *introspection.Field, tp *introspection.Type) (string, string, error) {
	name := fp.Name()
	typeName := *tp.Name()
	fieldCode := &bytes.Buffer{}
	methodCode := &bytes.Buffer{}

	fieldTypeName := g.getTypeName(fp.Type(), false)

	type fieldArgument struct {
		Name string
		Type string
	}
	fieldArguments := make([]fieldArgument, 0, len(fp.Args()))

	for _, field := range fp.Args() {
		fieldArguments = append(fieldArguments, fieldArgument{
			Name: field.Name(),
			Type: g.getTypeName(field.Type(), true),
		})
	}

	fieldTpl := map[string]interface{}{
		"TypeKind":         tp.Kind(),
		"FieldName":        name,
		"FieldDescription": args.removeLineBreaks(g.returnString(fp.Description())),
		"FieldType":        fieldTypeName,
	}

	// generate field template
	var err error
	fieldCode, err = args.ExecuteTemplateBuffer(GraphQLFieldTemplate, name, typeName, fieldTpl)
	if err != nil {
		return "", "", err
	}

	methodTpl := map[string]interface{}{
		"TypeKind":          tp.Kind(),
		"TypeName":          typeName,
		"MethodArguments":   fieldArguments,
		"MethodDescription": args.removeLineBreaks(g.returnString(fp.Description())),
		"MethodName":        name,
		"MethodReturnType":  fieldTypeName,
		"MethodReturn":      name,
	}

	// generate field template
	methodCode, err = args.ExecuteTemplateBuffer(GraphQLMethodTemplate, name, typeName, methodTpl)
	if err != nil {
		return "", "", err
	}

	return string(fieldCode.Bytes()), string(methodCode.Bytes()), nil
}

func (g *CodeGen) getPointer(typeName string, fp *introspection.Field) string {
	if fp.Type().Kind() == "NON_NULL" {
		return typeName
	}
	return "*" + typeName
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

func (g *CodeGen) isEntryPoint(a string) bool {
	return a == g.mutationName || a == g.queryName
}

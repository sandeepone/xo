package internal

import (
	//"strings"

	//"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/introspection"
)

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

func GenReturnType(t *Typ, mode, fieldName, pkg string) string {
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

	r += GenReturnType(t.Type, mode, fieldName, pkg)

	// Special case for edges - ugly hack
	if mode == "struct" && fieldName == "edges" {
		r = "*" + r
	}

	return r
}

// Resolvers generates the resolver function for the given FieldDef
func GenResolver(f *FieldDef, model, pkg string) string {
	var r string

	// res := GenReturnType(f.Type, "resolver", f.Name, pkg)
	// returnType := res

	// if f.Type.GQLType == "[]" {
	// 	if f.Type.IsNullable {
	// 		returnType = strings.Replace(returnType, "*", "", 1)
	// 	}
	// }

	if f.Type.GQLType != "graphql.ID" {
		r += "  return "

		if f.Type.IsNullable {
			r += "&"
		}
	}

	if _, ok := KnownGoTypes[f.Type.GQLType]; !ok {
		if f.Type.GQLType == "graphql.ID" {
			r += "  id := graphql.ID(r." + model + "." + f.Name + ")\n"
			r += "  return "

			if f.Type.IsNullable {
				r += "&"
			}

			r += "id"
		} else if f.Type.GQLType == "graphql.Time" {
			dref := ""
			if f.Type.IsNullable {
				dref = "*"
			}

			r += f.Type.GQLType + "{Time: " + dref + "r." + model + "." + f.Name + "}"
		} else {
			ref := ""
			if !f.Type.IsNullable {
				ref = "&"
			}

			r += f.Type.GQLType + "{" + ref + "r." + model + "." + f.Name + "}"
		}
	} else {
		r += "r." + model + "." + f.Name
	}

	return r
}

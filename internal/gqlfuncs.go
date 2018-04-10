package internal

import (
	"strings"

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

func GenType(t *Typ, mode, fieldName, pkg string) string {
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

	r += GenType(t.Type, mode, fieldName, pkg)

	// Special case for edges - ugly hack
	if mode == "struct" && fieldName == "edges" {
		r = "*" + r
	}

	return r
}

// Resolvers generates the resolver function for the given FieldDef
func GenResolver(f *FieldDef, isModel bool, pkg string) string {
	res := GenType(f.Type, "resolver", f.Name, pkg)
	//returnType := res

	r := "func (r *" + f.Parent + "Resolver) " + capitalise(f.Name) + "("
	r += ") " + res + " {\n"

	if !isModel {
		// Special Case for Edges
		if f.Name == "edges" {
			name := strings.TrimSuffix(f.Parent, "Connection")

			r += " if r." + name + "Connection.Edges == nil { \n"
			r += " return &[]*" + name + "EdgeResolver{} \n"
			r += " } \n\n"
		}

		// Special Case for PageInfo
		if f.Name == "pageInfo" {
			r += " if r." + f.Parent + ".PageInfo == nil { \n"
			r += " return &PageInfoResolver{PageInfo{}} \n"
			r += " } \n\n"
		}

		r += "  return "

		// Add pointer for Pageinfo cursors
		if f.Parent == "PageInfo" && (f.Name == "endCursor" || f.Name == "startCursor") {
			r += "&"
		}

		// if f.Type.IsNullable {
		// 	r += "&"
		// }

		r += "r." + f.Parent + "." + capitalise(f.Name)

		r += "\n}"
		return r
	}

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
			r += " if r." + f.Parent + "." + f.Name + " == \"\" { "
			r += "  return graphql.ID(\"\")"
			r += "  }\n\n"
			r += "//return graphql.ID(r." + f.Parent + "." + f.Name + ")\n"
			r += "  return "

			if f.Type.IsNullable {
				r += "&"
			}

			r += "relay.ToGlobalID(\"" + f.Parent + "\", r." + f.Parent + "." + f.Name + ")"
		} else if f.Type.GQLType == "graphql.Time" {
			dref := ""
			if f.Type.IsNullable {
				dref = "*"
			}

			r += f.Type.GQLType + "{Time: " + dref + "r." + f.Parent + "." + capitalise(f.Name) + "}"
		} else {
			ref := ""
			if !f.Type.IsNullable {
				ref = "&"
			}

			r += f.Type.GQLType + "{" + ref + "r." + f.Parent + "." + capitalise(f.Name) + "}"
		}
	} else {
		r += "r." + f.Parent + "." + capitalise(f.Name)
	}

	r += "\n}"
	return r
}

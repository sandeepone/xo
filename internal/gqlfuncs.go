package internal

import (
	//"log"
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

		// ok := KnownGoTypes[t.GoType]
		// if !ok {
		// 	r = t.GQLType
		// }
		//log.Printf("GenType code for %s %s", r, fieldName)
	} else {
		r = t.GQLType
	}

	if mode == "struct" {
		ok := KnownGoTypes[t.GoType]
		if !ok {
			r = t.GQLType
		}

		if t.IsNullable && t.GQLType != "[]" && !ok {
			r = "*" + r
		}
	} else {
		// if t.IsNullable {
		// 	r = "*" + r
		// }

		if mode == "interface" || mode == "query" {
			ok := KnownGoTypes[t.GoType]

			// Add golang package name if provided
			if pkg != "package" && pkg != "" && !ok {
				r = pkg + "." + r
			}

			// if t.IsNullable {
			// 	r = "*" + r
			// }
		}

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
	//itm := ""

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

		r += "r." + f.Parent + "." + capitalise(f.Name)
	} else {

		// if f.Type.GQLType == "[]" {
		// 	if f.Type.IsNullable {
		// 		returnType = strings.Replace(returnType, "*", "", 1)
		// 	}

		// 	if _, ok := KnownGoTypes[f.Type.Type.GoType]; !ok {
		// 		ref := ""
		// 		dref := ""

		// 		if f.Type.Type.IsNullable {
		// 			itm = "&"
		// 			dref = "*"
		// 		} else {
		// 			ref = "&"
		// 		}

		// 		if f.Type.Type.GQLType == "graphql.ID" {
		// 			itm = f.Type.Type.GQLType + "(itm)"
		// 		} else if f.Type.Type.GQLType == "graphql.Time" {
		// 			itm = itm + f.Type.Type.GQLType + "{Time: " + dref + "itm}"
		// 		} else {
		// 			itm = itm + f.Type.Type.GQLType + "{" + ref + "itm}"
		// 		}
		// 	} else {
		// 		if f.Type.Type.IsNullable {
		// 			itm = "&itm"
		// 		}
		// 	}

		// 	if itm == "" {
		// 		itm = "itm"
		// 	}
		// 	r += "  items := " + returnType + "{}\n"
		// 	r += "  for _, itm := range r.R." + f.Name + " {\n"
		// 	r += "    items = append(items, " + itm + ")\n"
		// 	r += "  }\n"
		// 	r += "  return "
		// 	if f.Type.IsNullable {
		// 		r += "&"
		// 	}

		// 	r += "items\n"
		// 	r += "}"
		// 	return r
		// }

		if f.Type.GQLType != "graphql.ID" {
			r += "  return "

			if f.Type.IsNullable {
				r += "&"
			}
		}

		if _, ok := KnownGoTypes[f.Type.GQLType]; !ok {
			//log.Printf("GenResolver code for %s %s", f.Type.GQLType, f.Name)

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
				// if f.Type.IsNullable {
				// 	dref = "*"
				// }

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
	}

	r += "\n}"
	return r
}

func GenInterface(t *TypeDef, mode, pkg string) string {
	r := "type " + t.Name + " interface {\n"

	for _, f := range t.Fields {
		r += "  " + capitalise(f.Name) + "("
		if len(f.Args) > 0 {
			r += capitalise(f.Name) + "Args"
		}

		r += ") (" + GenType(f.Type, "interface", f.Name, pkg) + ", error) \n"
	}

	r += "}"
	return r
}

func GenFuncArgs(f *FieldDef, mode, pkg string) string {
	r := "type " + capitalise(f.Name) + "Args struct {\n"

	for _, arg := range f.Args {
		r += "  " + arg.Name + " " + GenType(arg.Type, "argStruct", arg.Name, pkg) + "\n"
	}

	r += "}"
	return r
}

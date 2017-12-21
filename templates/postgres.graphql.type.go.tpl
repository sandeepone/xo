{{if eq .Kind "OBJECT"}}
{{if not (is_entry .TypeName) }}
// {{.TypeName}} {{.TypeDescription}} - OBJECT
type {{.TypeName}} struct {
  {{range .Fields}}{{.}}{{end}}
}

// {{.TypeName}}Resolver resolver for {{.TypeName}} - Type Resolver
type {{.TypeName}}Resolver struct {
  {{.TypeName}}
}
{{end}}
{{range .Methods}}{{.}}
{{end}}
{{if not (is_entry .TypeName) }}
func (r *{{.TypeName}}Resolver) MarshalJSON() ([]byte, error) {
  return json.Marshal(&r.{{.TypeName}})
}

func (r *{{.TypeName}}Resolver) UnmarshalJSON(data []byte) error {
  return json.Unmarshal(data, &r.{{.TypeName}})
}
{{end}}
{{end}}

{{if eq .Kind "INTERFACE"}}
// {{.TypeName}} {{.TypeDescription}} - INTERFACE
type {{.TypeName}} interface {
  {{range .Methods}}{{.}}{{end}}
}

// {{.TypeName}}Resolver resolver for {{.TypeName}}
type {{.TypeName}}Resolver struct {
  {{.TypeName}}
}
{{ $typeName := .TypeName }}
{{range $possibleType := .PossibleTypes}}
  func (r *{{$typeName}}Resolver) To{{$possibleType}}() (*{{$possibleType}}Resolver, bool) {
    c, ok := r.{{$typeName}}.(*{{$possibleType}}Resolver)
	   return c, ok
  }
{{end}}

{{end}}

{{if eq .Kind "UNION"}}
// {{.TypeName}}Resolver resolver for {{.TypeName}} - UNION
type {{.TypeName}}Resolver struct {
  {{.TypeName | uncapitalize}} interface{}
}
{{ $typeName := .TypeName }}
{{range $possibleType := .PossibleTypes}}
  func (r *{{$typeName}}Resolver) To{{$possibleType}}() (*{{$possibleType}}Resolver, bool) {
    c, ok := r.{{$typeName | uncapitalize}}.(*{{$possibleType}}Resolver)
	   return c, ok
  }
{{end}}

{{end}}

{{if eq .Kind "ENUM"}}
{{ $typeName := .TypeName }}
{{ $typeDescription := .TypeDescription }}
// {{.TypeName}} {{.TypeDescription}} - ENUM
type {{$typeName}} string
const (
{{range $value := .EnumValues}}
  // {{$typeName}}{{$value}} {{$typeDescription}}
  {{$typeName}}{{$value}} = {{$typeName}}("{{$value}}")
{{end}}
)
{{end}}

{{if eq .Kind "INPUT_OBJECT"}}
// {{.TypeName}} {{.TypeDescription}} INPUT_OBJECT
type {{.TypeName}} struct {
  {{range .InputFields}}{{.}}{{end}}
}
{{end}}

{{if eq .Kind "RESOLVER"}}
// {{.TypeName}} {{.TypeDescription}} - RESOLVER
type {{.TypeName}} struct {
}
{{end}}

{{if eq .Kind "SCALAR"}}
// {{.TypeName}}Resolver {{.TypeDescription}} - SCALAR
type {{.TypeName}}Resolver struct {
  value interface{}
}

func (r *{{.TypeName}}Resolver) ImplementsGraphQLType(name string) bool {
    return false
}

func (r *{{.TypeName}}Resolver) UnmarshalGraphQL(input interface{}) error {
  // Scalars need to be implemented manually
  r.value = input
  return nil
}

{{end}}

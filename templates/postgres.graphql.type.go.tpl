{{if eq .Kind "OBJECT"}}

// {{.TypeName}} {{.TypeDescription}}
type {{.TypeName}} struct {
  {{range .Fields}}{{.}}{{end}}
}

// {{.TypeName}}Resolver resolver for {{.TypeName}} - Type Resolver
type {{.TypeName}}Resolver struct {
  {{.TypeName}}
}

{{range .Methods}} {{.}} {{end}}

func (r *{{.TypeName}}Resolver) MarshalJSON() ([]byte, error) {
  return json.Marshal(&r.{{.TypeName}})
}

func (r *{{.TypeName}}Resolver) UnmarshalJSON(data []byte) error {
  return json.Unmarshal(data, &r.{{.TypeName}})
}

{{end}}

{{if eq .Kind "INTERFACE"}}
// {{.TypeName}} {{.TypeDescription}}
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
// {{.TypeName}}Resolver resolver for {{.TypeName}}
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
// {{.TypeName}} {{.TypeDescription}}
type {{$typeName}} string
const (
{{range $value := .EnumValues}}
  // {{$typeName}}{{$value}} {{$typeDescription}}
  {{$typeName}}{{$value}} = {{$typeName}}("{{$value}}")
{{end}}
)
{{end}}

{{if eq .Kind "INPUT_OBJECT"}}
// {{.TypeName}} {{.TypeDescription}}
type {{.TypeName}} struct {
  {{range .InputFields}}{{.}}{{end}}
}
{{end}}

{{if eq .Kind "RESOLVER"}}
// {{.TypeName}} {{.TypeDescription}}
type {{.TypeName}} struct {
}
{{end}}

{{if eq .Kind "SCALAR"}}
// {{.TypeName}}Resolver {{.TypeDescription}}
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

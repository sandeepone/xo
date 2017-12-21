{{if eq .Kind "OBJECT"}}

{{range .Methods}} {{.}} {{end}}

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

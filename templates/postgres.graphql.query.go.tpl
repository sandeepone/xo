{{ define "field" }}
// {{capitalize .FieldName}} {{.FieldDescription}}
{{capitalize .FieldName}} {{.FieldType}} `json:"{{.FieldName}}"`
{{- end }}

{{- define "method" }}
{{if eq .TypeKind "OBJECT"}}
{{$hasArguments := gt (.MethodArguments | len) 0}}
{{if $hasArguments}}
// {{capitalize .MethodName}}QueryArgs are the arguments for the "{{.MethodName}}" query.
type {{capitalize .MethodName}}QueryArgs struct {
  {{range .MethodArguments}} {{.Name | capitalize}} {{.Type}}
  {{end}}
}
{{end}}
// {{capitalize .MethodName}} {{.MethodDescription}} - Method
func (r *Resolver) {{capitalize .MethodName}}(ctx context.Context{{if $hasArguments}}, args {{capitalize .MethodName}}QueryArgs{{end}}) ({{.MethodReturnType}}, error) {

  return nil, nil
}
{{- end}}

{{if eq .TypeKind "INTERFACE"}}
{{$hasArguments := gt (.MethodArguments | len) 0}}
// {{capitalize .MethodName}} {{.MethodDescription}}
{{capitalize .MethodName}}({{if $hasArguments}}{{template "arguments" .MethodArguments}}{{end}}) {{.MethodReturnType}}
{{end}}

{{- end }}



{{if eq .Kind "OBJECT"}}
{{range .Methods}} {{template "method" .}} {{end}}
{{end}}

{{if eq .Kind "INTERFACE"}}
// {{.TypeName}} {{.TypeDescription}} - INTERFACE
type {{.TypeName}} interface {
  {{range .Methods}} {{template "method" .}} {{end}}
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
  {{range .InputFields}} {{template "field" .}} {{end}}
}
{{end}}

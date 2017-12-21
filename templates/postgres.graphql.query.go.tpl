{{ define "arguments" }}args *struct{
  {{range .}}{{.Name | capitalize}} {{.Type}}
  {{end}}
}{{- end }}

{{- define "receiver"}} {{if .IsEntry }}Resolver{{else}}{{.TypeName}}Resolver{{end}}
{{- end}}

{{ define "field" }}
// {{capitalize .FieldName}} {{.FieldDescription}}
{{capitalize .FieldName}} {{.FieldType}} `json:"{{.FieldName}}"`
{{- end }}

{{- define "method" }}
{{if eq .TypeKind "OBJECT"}}
{{$hasArguments := gt (.MethodArguments | len) 0}}
// {{capitalize .MethodName}} {{.MethodDescription}} - Method
func (r *{{template "receiver" .}}) {{capitalize .MethodName}}({{if $hasArguments}}{{template "arguments" .MethodArguments}}{{end}}) {{.MethodReturnType}} {
  {{- if .IsEntry}}
  return nil
  {{- else}}
  return r.{{.TypeName}}.{{capitalize .MethodReturn}}
  {{- end}}
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

{{if eq .Kind "RESOLVER"}}
// {{.TypeName}} {{.TypeDescription}} - RESOLVER
type {{.TypeName}} struct {
}
{{end}}

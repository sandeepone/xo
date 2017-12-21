{{ define "arguments" }}args *struct{
  {{range .}}{{.Name | capitalize}} {{.Type}}
  {{end}}
}{{ end }}

{{define "receiver"}} {{if .IsEntry }}Resolver{{else}}{{.TypeName}}Resolver{{end}} {{end}}

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

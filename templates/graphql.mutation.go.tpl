{{- define "inputField" }}
    // {{capitalize .Name}} {{.Description}}
    {{capitalize .Name}} {{.FieldType}} `json:"{{.Name}}"`
{{- end }}

{{- define "payloadField" }}
    // {{capitalize .Name}} {{.Description}}
    {{lcfirst .Name}} *objects.{{.NFieldType}} `json:"{{.Name}}"`
{{- end }}

{{- define "arguments" -}} 
  args *struct{ {{ range . }} {{.Name | capitalize }} {{.Type }}
  {{ end }} }
{{- end }}

{{- define "returnType" }}
    {{- if .IsMutation }} (*{{.NReturnType}}, error) {{- else }} *objects.{{.NReturnType}} {{- end }}
{{- end }}



{{- define "method" }}
    {{- if eq .TypeKind "OBJECT" }}
        {{$hasArguments := gt (.Arguments | len) 0}}

        // {{capitalize .Name}} {{.Description}}
        func (r *{{.TypeName}}) {{capitalize .Name}}({{- if $hasArguments }} {{template "arguments" .Arguments}} {{- end}}) {{ template "returnType" .}} {
              
          return {{- if .IsMutation }} nil,nil {{- else }} r.{{lcfirst .Return}} {{- end }}
        }
    {{- end }}
{{- end }}




{{if eq .Kind "INPUT_OBJECT"}}
    // {{.TypeName}} {{.TypeDescription}} INPUT_OBJECT
    type {{.TypeName}} struct {
      {{- range .InputFields}} {{template "inputField" .}} {{end}}
    }
{{end}}



{{if eq .Kind "OBJECT"}}
    {{if ne .TypeName "Mutation"}}
      // {{.TypeName}} {{.TypeDescription}}
      type {{.TypeName}} struct {
        {{- range .Fields}} {{template "payloadField" .}} {{- end}}
      }
    {{ end}}

    {{range .Methods}} {{template "method" .}} {{end}}
{{ end}}



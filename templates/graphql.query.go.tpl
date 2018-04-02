{{- define "field" }}
    // {{capitalize .Name}} {{.Description}}
    {{capitalize .Name}} {{.FieldType}} `json:"{{.Name}}"`
{{- end }}

{{- define "returnType" }}
    {{- if .IsNullable }} *objects.{{.NReturnType}} {{- else }} objects.{{.NReturnType}} {{- end }}
{{- end }}

{{- define "method" }}
    {{- if eq .TypeKind "OBJECT" }}
        {{$hasArguments := gt (.Arguments | len) 0}}

        {{- if $hasArguments }}
            // {{capitalize .Name}}Args are the arguments for the "{{.Name}}" query.
            type {{capitalize .Name}}Args struct {
              {{range .Arguments}} {{.Name | capitalize}} {{.Type}}
              {{end}}
            }
        {{- end }}

        // {{capitalize .Name}} {{.Description}} - {{.ReturnType}}
        func (r *{{.TypeName}}Resolver) {{capitalize .Name}}(ctx context.Context {{- if $hasArguments }}, args {{capitalize .Name}}Args{{- end }} ) (*objects.{{.NReturnType}}, error) {

          return nil, nil
        }
    {{- end }}

    {{- if eq .TypeKind "INTERFACE" }}
        {{$hasArguments := gt (.Arguments | len) 0}}
        // {{capitalize .Name}} {{.Description}}
        {{capitalize .Name}}({{if $hasArguments}}{{template "arguments" .Arguments}}{{end}}) {{.ReturnType}}
    {{- end }}
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

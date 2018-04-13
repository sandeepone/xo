
{{- if eq .Template "MUTATION" }}

    // {{capitalize .Name}} - {{.Description}}
    type {{.Name}} interface {
      {{- range .Fields}}
        {{ $hasArguments := gt (.Args | len) 0}}
        // {{capitalize .Name}} {{.Description}}
        {{capitalize .Name}}({{if $hasArguments}}context.Context, {{capitalize .Name}}Args{{end}}) ({{genType .Type "interface" .Name "package"}}, error)
      {{- end}}
    }


{{- end}}




{{- if eq .Template "INPUT" }}

  // {{.Name}} - {{.Description}}
  type {{.Name}} struct {
    {{- range .Fields}}
      {{capitalize .Name}} {{genType .Type "struct" .Name "package"}}  `json:"{{.Name}}"` {{- if ne .Description "" }} // {{.Description}} {{- end}}
    {{- end}}
  }

{{- end}}



{{- if eq .Template "PAYLOAD" }}

  {{- $TypeName := (.Name) -}}
  {{- $IsModel := (.IsModel) -}}

  // {{.Name}} {{.Description}}
  type {{.Name}} struct {
    {{- range .Fields}}
      {{capitalize .Name}} {{genType .Type "struct" .Name "objects"}}  `json:"{{.Name}}"` // {{capitalize .Name}} {{.Description}}
    {{- end}}
  }
  
  // {{.Name}}Resolver resolver for {{.Name}}
  type {{.Name}}Resolver struct {
    {{.Name}}
  }
  
  {{ range .Fields}}
    // {{capitalize .Name}} {{.Description}}
    func (r *{{capitalize .Parent}}Resolver) {{capitalize .Name}}() ({{genType .Type "interface" .Name "objects"}}, error) {
      return r.{{capitalize .Parent}}.{{capitalize .Name}}
    }
    
  {{ end}}

{{- end}}



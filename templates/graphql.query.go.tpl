
{{- if eq .Template "QUERY" }}

  // {{capitalize .Name}} - {{.Description}}
  type {{capitalize .Name}}Resolver interface {
    {{- range .Fields}}
      {{ $hasArguments := gt (.Args | len) 0}}
      // {{capitalize .Name}} {{.Description}}
      {{capitalize .Name}}({{if $hasArguments}}ctx context.Context, args {{capitalize .Name}}Args{{end}}) ({{genType .Type "interface" .Name "objects"}}, error)
    {{- end}}
  }
  
  {{- range .Fields}}
    // {{capitalize .Name}}Args are the arguments for the "{{capitalize .Name}}" query.
    type {{capitalize .Name}}Args struct{
      {{- range .Args}}
        {{capitalize .Name}} {{genType .Type "argStruct" .Name "package"}}
      {{- end}}
    }
    
  {{- end}}

{{- end}}




{{- if eq .Template "MUTATION" }}

  // {{capitalize .Name}} - {{.Description}}
  type {{capitalize .Name}}Resolver interface {
    {{- range .Fields}}
      {{- $hasArguments := gt (.Args | len) 0}}
      
      // {{capitalize .Name}} {{.Description}}
      {{capitalize .Name}}({{if $hasArguments}}ctx context.Context, args *{{capitalize .Name}}Args{{end}}) ({{genType .Type "interface" .Name "package"}}, error)
    {{- end}}
  }

{{- end}}



{{- if eq .Template "ARGS" }}

  {{- range .Fields}}
    // {{capitalize .Name}}Args are the arguments for the "{{capitalize .Name}}" query.
    type {{capitalize .Name}}Args struct{
      {{- range .Args}}
        {{capitalize .Name}} {{genType .Type "argStruct" .Name "package"}}
      {{- end}}
    }
    
  {{- end}}
  
{{- end}}


  
{{- if eq .Template "INPUT" }}

  // {{.Name}} - {{.Description}}
  type {{.Name}} struct {
    {{- range .Fields}}
      {{capitalize .Name}} {{genType .Type "argStruct" .Name "package"}}  `json:"{{.Name}}"` {{- if ne .Description "" }} // {{.Description}} {{- end}}
    {{- end}}
  }

{{- end}}



{{- if eq .Template "PAYLOAD" }}

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

  // New{{.Name}} for {{.Name}}
  func New{{.Name}}(pl {{.Name}}) *{{.Name}}Resolver {
    return &{{.Name}}Resolver{pl}
  }
  
  {{ range .Fields}}
    // {{capitalize .Name}} {{.Description}}
    func (r *{{capitalize .Parent}}Resolver) {{capitalize .Name}}() {{genType .Type "interface" .Name "objects"}} {
      return r.{{capitalize .Parent}}.{{capitalize .Name}}
    }
    
  {{ end}}

{{- end}}




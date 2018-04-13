
// {{capitalize .Name}} - {{.Description}}
type {{.Name}} interface {
  {{- range .Fields}}
    {{ $hasArguments := gt (.Args | len) 0}}
    // {{capitalize .Name}} {{.Description}}
    {{capitalize .Name}}({{if $hasArguments}}context.Context, {{capitalize .Name}}Args{{end}}) (*{{genType .Type "interface" .Name "objects"}}, error)
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

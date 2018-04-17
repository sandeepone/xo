{{- $TypeName := (.Name) -}}
{{- $IsModel := (.IsModel) -}}



{{- if eq .Template "TYPE" }}
  // {{.Name}} {{.Description}}
  type {{.Name}} struct {
    {{- range .Fields}}
      {{- if ne .Name "nid" }}
        {{capitalize .Name}} {{genType .Type "struct" .Name "package"}}  `json:"{{.Name}}"` // {{capitalize .Name}} {{.Description}}
      {{- end}}
    {{- end}}
  }
  
  // {{.Name}}Resolver resolver for {{.Name}}
  type {{.Name}}Resolver struct {
    {{.Name}}
  }
  
  {{ range .Fields}}
    // {{capitalize .Name}} {{.Description}}
    {{ genResolver . $IsModel "package"}}
  {{ end}}
  
  func (r *{{.Name}}Resolver) MarshalJSON() ([]byte, error) {
    return json.Marshal(&r.{{.Name}})
  }
  
  func (r *{{.Name}}Resolver) UnmarshalJSON(data []byte) error {
    return json.Unmarshal(data, &r.{{.Name}})
  }
{{- end}}




{{- if eq .Template "EXTRA" }}

  {{ range .Efields}}
    // {{capitalize .Name}} - {{.Description}}
    {{ genResolver . $IsModel "package"}}
  {{ end}}
  
{{- end}}


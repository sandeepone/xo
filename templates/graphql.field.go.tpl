{{ define "field" }}
// {{capitalize .FieldName}} {{.FieldDescription}}
{{capitalize .FieldName}} {{.FieldType}} `json:"{{.FieldName}}"`
{{- end }}

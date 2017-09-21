{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- $stable := (.Table.TableName) -}}

var {{ .Name }}Type = graphql.NewObject(graphql.ObjectConfig{
			Name: "{{ .Name }}",
			{{- if .Comment }}
			Description: "{{ .Comment }}.",
			{{- else }}
			Description: "The {{.Name}} represents a row from '{{ $stable }}'.",
			{{- end }}
			Fields: graphql.Fields{
				{{- range .Fields }}
					"{{- .Col.ColumnName }}": &graphql.Field{ Type: graphql.{{ gqltype .Type }}, },
				{{- end}}
			},
})

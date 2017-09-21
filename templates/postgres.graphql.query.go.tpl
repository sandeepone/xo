{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- $stable := (.Table.TableName) -}}

	var {{ $stable }}Type = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "{{ $stable }}",
			Fields: graphql.Fields{
				{{- range .Fields }}
					"{{- .Col.ColumnName }}": &graphql.Field{
						Type: graphql.{{ retype .Type }},
					},
				{{- end}}
			},
		},
	)

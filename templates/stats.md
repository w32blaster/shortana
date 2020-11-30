Statistics grouped by Short URLs:

{{ range $key, $value := .Stats }}
   - {{$.Hostname}}/{{.ShortUrl}} Total views: {{ .TotalViews }} (with {{ .TotalUniqueUsers }} unique users) for {{ .TotalDaysActive }} days since {{ .PublishDate}} /stats{{ .ID }}
{{ end }}
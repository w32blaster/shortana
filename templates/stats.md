Statistics grouped by Short URLs:

{{ range $key, $value := .Stats }}
   o `{{.ShortUrl}}` Total views: {{ .TotalViews }} (with {{ .TotalUniqueUsers }} unique users) for {{ .TotalDaysActive }} days since `{{ .PublishDate}}` /stats{{ .ShortUrlID }}
{{ end }}
Statistics grouped by Short URLs:

{{ range $key, $value := .Stats }}
  \- [{{ markdownEscape .ShortUrl}}]({{$.Hostname}}/{{.ShortUrl}})\. Total views: {{ .TotalViews }} \(with {{ .TotalUniqueUsers }} unique users\) for {{ .TotalDaysActive }} days since {{ markdownEscape .PublishDate}} /stats{{ .ShortUrlID }}
{{ end }}
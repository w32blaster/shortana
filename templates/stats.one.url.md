Full view statistics for {{ markdownEscape .ShortURL.ShortUrl }}:

{{ range $key, $value := .Stats }}
 \- {{ markdownEscape $key }}: {{ .TotalViews }} views \({{ .UniqueViews }} unique\) /stats{{ $.ShortURL.ID }}x{{ .DateWithoutHyphens }}
{{ end }}

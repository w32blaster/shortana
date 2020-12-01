Full view statistics for {{ .ShortURL.ShortUrl }}:

{{ range $key, $value := .Stats }}
 \- {{ markdownEscape $key }}: {{ .TotalViews }} views \({{ .UniqueViews }} unique\) /stats{{ markdownEscape $.ShortURL.ID }}x{{ .DateWithoutHyphens }}
{{ end }}

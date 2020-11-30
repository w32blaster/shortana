Full view statistics for {{ .ShortURL.ShortUrl }}:

{{ range $key, $value := .Stats }}
 - {{ $key }}: {{ .TotalViews }} views ({{ .UniqueViews }} unique) /stats{{$.ShortURL.ID}}x{{ .DateWithoutHyphens }}
{{ end }}

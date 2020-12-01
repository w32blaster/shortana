Full view statistics for {{ .ShortURL.ShortUrl }}:

{{ range $key, $value := .Stats }}
 x `{{ $key }}`: {{ .TotalViews }} views ({{ .UniqueViews }} unique) /stats{{$.ShortURL.ID}}x{{ .DateWithoutHyphens }}
{{ end }}

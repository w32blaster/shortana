Views statistics for {{ .ShortURL.ShortUrl }} at {{ .ShortURL.PublishDate }}:

{{ range .Views }} {{ $length := len .ViewTimes }}
 x `{{ .UserIpAddress }}` {{ $length }} views from {{ .City}} ({{ .CountryCode }}); User-Agent: {{ .UserAgent }}
{{ end }}

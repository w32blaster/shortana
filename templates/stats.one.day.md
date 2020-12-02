Views statistics for {{ markdownEscape .ShortURL.ShortUrl }} at {{ markdownEscape .SelectedDate }}:
{{ range .Views }} {{ $length := len .ViewTimes }}
 \- {{ markdownEscape .UserIpAddress }} {{ $length }} views from {{ markdownEscape .City}} \({{ markdownEscape .CountryCode }}\); /view{{ .ID }}
{{ end }}

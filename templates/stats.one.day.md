Views statistics for {{ markdownEscape .ShortURL.ShortUrl }} at {{ markdownEscape .ShortURL.PublishDate }}:

{{ range .Views }} {{ $length := len .ViewTimes }}
 \- {{ markdownEscape .UserIpAddress }} {{ $length }} views from {{ markdownEscape .City}} \({{ markdownEscape .CountryCode }}\); User\-Agent: {{ markdownEscape .UserAgent }}
{{ end }}

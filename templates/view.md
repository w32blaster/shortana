*One View stats for {{ markdownEscape .ShortUrl }} at {{ markdownEscape .Day }}*

ID: {{ .ID}}
IP: {{ markdownEscape .UserIpAddress }}
Country: {{ markdownEscape .CountryName}} \({{ markdownEscape .CountryCode }}\)
City: {{ markdownEscape .City }}
UA: {{ markdownEscape .UserAgent }}
Views count: {{ len .ViewTimes }}
Views Times:
{{ range .ViewTimes }}
 \- {{ formatDate . }} 
{{ end }}
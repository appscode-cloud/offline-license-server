# {{ .ProductLine }} {{ .Release }}

{{ range $p := .Projects }}
## [{{ trimPrefix "github.com/" $p.URL }}](https://{{ $p.URL }})
{{ range $r := $p.Releases }}
### [{{ $r.Tag }}](https://{{ $p.URL }}/releases/tag/{{ $r.Tag }})

{{ range $c := $r.Commits -}}
 - [{{ substr 0 8 $c.SHA }}](https://{{ $p.URL }}/commit/{{ $c.SHA }}) {{ $c.Subject }}
{{ end }}
{{ end }}
{{ end }}

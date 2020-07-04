# {{ .ProductLine }} Releases

| {{ .ProductLine }} Version | Release Date | User Guide | Changelog | Kubernetes Version |
|--------------------------- | ------------ | ---------- | --------- | ------------------ |
{{ range $r :=  .Releases -}}
| [{{ $r.Release }}]({{ $r.ReleaseURL }}) | {{ $r.ReleaseDate | date "2006-01-02" }} | [User Guide]({{ $r.DocsURL }}) | [CHANGELOG]({{ $r.ChangelogURL }}) | {{ $r.KubernetesVersion }} |
{{ end }}

# Change Log

All notable changes to this project will be documented in this file.

{{ if .Versions -}}
<a name="unreleased"></a>
## [Unreleased]
{{ if .Unreleased.CommitGroups -}}
{{ range .Unreleased.CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
{{/* SKIPPING RULES - START */ -}}
{{- if not (hasPrefix .Subject "Updated CHANGELOG") -}}
{{- if not (contains .Subject "[ci skip]") -}}
{{- if not (contains .Subject "[skip ci]") -}}
{{- if not (hasPrefix .Subject "Merge pull request ") -}}
{{- if not (hasPrefix .Subject "Added CHANGELOG") -}}
{{- /* SKIPPING RULES - END */ -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{/* SKIPPING RULES - START */ -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{/* SKIPPING RULES - END */ -}}
{{ end }}
{{ end -}}
{{ else }}
{{ range .Unreleased.Commits -}}
{{/* SKIPPING RULES - START */ -}}
{{- if not (hasPrefix .Subject "Updated CHANGELOG") -}}
{{- if not (contains .Subject "[ci skip]") -}}
{{- if not (contains .Subject "[skip ci]") -}}
{{- if not (hasPrefix .Subject "Merge pull request ") -}}
{{- if not (hasPrefix .Subject "Added CHANGELOG") -}}
{{- /* SKIPPING RULES - END */ -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{/* SKIPPING RULES - START */ -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{/* SKIPPING RULES - END */ -}}
{{ end }}
{{ end -}}
{{ end -}}

{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]{{ else }}{{ .Tag.Name }}{{ end }} - {{ datetime "2006-01-02" .Tag.Date }}
{{ if .CommitGroups -}}
{{ range .CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
{{/* SKIPPING RULES - START */ -}}
{{- if not (hasPrefix .Subject "Updated CHANGELOG") -}}
{{- if not (contains .Subject "[ci skip]") -}}
{{- if not (contains .Subject "[skip ci]") -}}
{{- if not (hasPrefix .Subject "Merge pull request ") -}}
{{- if not (hasPrefix .Subject "Added CHANGELOG") -}}
{{- /* SKIPPING RULES - END */ -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{/* SKIPPING RULES - START */ -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{/* SKIPPING RULES - END */ -}}
{{ end }}
{{ end -}}
{{ else }}
{{ range .Commits -}}
{{/* SKIPPING RULES - START */ -}}
{{- if not (hasPrefix .Subject "Updated CHANGELOG") -}}
{{- if not (contains .Subject "[ci skip]") -}}
{{- if not (contains .Subject "[skip ci]") -}}
{{- if not (hasPrefix .Subject "Merge pull request ") -}}
{{- if not (hasPrefix .Subject "Added CHANGELOG") -}}
{{- /* SKIPPING RULES - END */ -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{/* SKIPPING RULES - START */ -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{/* SKIPPING RULES - END */ -}}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### {{ .Title }}
{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}

{{- if .Versions }}
[Unreleased]: {{ .Info.RepositoryURL }}/compare/{{ $latest := index .Versions 0 }}{{ $latest.Tag.Name }}...HEAD
{{ range .Versions -}}
{{ if .Tag.Previous -}}
[{{ .Tag.Name }}]: {{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}
{{ end -}}
{{ end -}}
{{ end -}}

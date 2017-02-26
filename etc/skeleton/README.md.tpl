# {{.Manifest.Language}}-driver  ![Driver Status](https://img.shields.io/badge/status-{{.Manifest.Status | escape_shield}}-{{template "color-status" .}}.svg) [![Build Status](https://travis-ci.org/bblfsh/{{.Manifest.Language}}-driver.svg?branch=master)](https://travis-ci.org/bblfsh/{{.Manifest.Language | escape_shield }}-driver) ![Native Version](https://img.shields.io/badge/{{.Manifest.Language}}%20version-{{.Manifest.Runtime.NativeVersion | escape_shield}}-aa93ea.svg) ![Go Version](https://img.shields.io/badge/go%20version-{{.Manifest.Runtime.GoVersion | escape_shield}}-63afbf.svg)

{{.Manifest.Documentation.Description}}

{{if .Manifest.Documentation.Caveats -}}
Caveats
-------

{{.Manifest.Documentation.Caveats}}
{{end -}}

License
-------

GPLv3, see [LICENSE](LICENSE)


{{define "color-status" -}}
{{if eq .Manifest.Status "planning" -}}
e08dd1
{{- else if eq .Manifest.Status "pre-alpha" -}}
d6ae86
{{- else if eq .Manifest.Status "alpha" -}}
db975c
{{- else if eq .Manifest.Status "beta" -}}
dbd25c
{{- else if eq .Manifest.Status "stable" -}}
9ddb5c
{{- else if eq .Manifest.Status "mature" -}}
60db5c
{{- else -}}
d1d1d1
{{- end}}
{{- end}}
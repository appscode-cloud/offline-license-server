{{- $version := semver .Release -}}
---
title: Changelog | {{ .ProductLine }}
description: Changelog
menu:
  docs_{{ "{{.version}}" }}:
    identifier: changelog-{{ lower .ProductLine }}-{{ .Release }}
    name: Changelog-{{ .Release }}
    parent: welcome
    weight: {{printf "%d%02d%02d" $version.Major $version.Minor $version.Patch}}
product_name: {{ lower .ProductLine }}
menu_name: docs_{{ "{{.version}}" }}
section_menu_id: welcome
url: /docs/{{ "{{.version}}" }}/welcome/changelog-{{ .Release }}/
aliases:
  - /docs/{{ "{{.version}}" }}/CHANGELOG-{{ .Release }}/
---

{{template "changelog.tpl" .}}

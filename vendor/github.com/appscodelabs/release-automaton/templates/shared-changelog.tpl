{{- $version := semver .Release -}}
---
title: Changelog | {{ .ProductLine }}
description: Changelog
menu:
  product_{{ lower .ProductLine }}_{{ "{{.version}}" }}:
    identifier: changelog-{{ lower .ProductLine }}
    name: Changelog
    parent: welcome
    weight: {{printf "%d%02d%02d" $version.Major $version.Minor $version.Patch}}
product_name: {{ lower .ProductLine }}
menu_name: product_{{ lower .ProductLine }}_{{ "{{.version}}" }}
section_menu_id: welcome
url: /products/{{ lower .ProductLine }}/{{ "{{.version}}" }}/welcome/changelog/
aliases:
  - /products/{{ lower .ProductLine }}/{{ "{{.version}}" }}/CHANGELOG/
---

{{template "changelog.tpl" .}}

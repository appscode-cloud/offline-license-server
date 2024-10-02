[![Build Status][travis-image]][travis-url]
[![Github Tag][githubtag-image]][githubtag-url]

[![Coverage Status][coveralls-image]][coveralls-url]
[![Maintainability][codeclimate-image]][codeclimate-url]

[![Go Report Card][goreport-image]][goreport-url]
[![GoDoc][godoc-image]][godoc-url]
[![License][license-image]][license-url]

***

# go-diacritics

> A package to handle diacritics

Provides a method to remove diacritical characters from any string and
replace them with their ASCII representation.

It handles all cases where an unicode decomposition exists (e.g. ä and è) as
well as all known latin cases without an unicode decomposition as listed below.

## Usage

To get the lastest tagged version of package, execute:

```
go get gopkg.in/Regis24GmbH/go-diacritics.v2
```

To import this package, add the following line to your code:

```
import "gopkg.in/Regis24GmbH/go-diacritics.v2"
```

This is a code example:

```
func main() {
  noDiacrits := godiacritics.Normalize("än éᶍample")
  println(noDiacrits) // prints "an example"
}
```

***

[travis-image]: https://travis-ci.org/Regis24GmbH/go-diacritics.svg?branch=master
[travis-url]: https://travis-ci.org/Regis24GmbH/go-diacritics

[githubtag-image]: https://img.shields.io/github/tag/Regis24GmbH/go-diacritics.svg?style=flat
[githubtag-url]: https://github.com/Regis24GmbH/go-diacritics

[coveralls-image]: https://coveralls.io/repos/github/Regis24GmbH/go-diacritics/badge.svg?branch=master
[coveralls-url]: https://coveralls.io/github/Regis24GmbH/go-diacritics?branch=master

[codeclimate-image]: https://api.codeclimate.com/v1/badges/91b466506779e639b614/maintainability
[codeclimate-url]: https://codeclimate.com/github/Regis24GmbH/go-diacritics/maintainability

[goreport-image]: https://goreportcard.com/badge/github.com/Regis24GmbH/go-diacritics
[goreport-url]: https://goreportcard.com/report/github.com/Regis24GmbH/go-diacritics

[godoc-image]: https://godoc.org/github.com/Regis24GmbH/go-diacritics?status.svg
[godoc-url]: https://godoc.org/github.com/Regis24GmbH/go-diacritics

[license-image]: https://img.shields.io/github/license/Regis24GmbH/go-diacritics.svg?style=flat
[license-url]: https://github.com/Regis24GmbH/go-diacritics/blob/master/LICENSE

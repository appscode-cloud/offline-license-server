# cache

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/go-macaron/cache/Go?logo=github&style=for-the-badge)](https://github.com/go-macaron/cache/actions?query=workflow%3AGo)
[![codecov](https://img.shields.io/codecov/c/github/go-macaron/cache/master?logo=codecov&style=for-the-badge)](https://codecov.io/gh/go-macaron/cache)
[![GoDoc](https://img.shields.io/badge/GoDoc-Reference-blue?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/go-macaron/cache?tab=doc)
[![Sourcegraph](https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?style=for-the-badge&logo=sourcegraph)](https://sourcegraph.com/github.com/go-macaron/cache)

Middleware cache provides cache management for [Macaron](https://github.com/go-macaron/macaron). It can use many cache adapters, including memory, file, Redis, Memcache, PostgreSQL, MySQL, Ledis and Nodb.

### Installation

The minimum requirement of Go is 1.6 (*1.7 if using Redis, 1.10 if using MySQL*).

	go get github.com/go-macaron/cache

## Getting Help

- [API Reference](https://gowalker.org/github.com/go-macaron/cache)
- [Documentation](https://go-macaron.com/middlewares/cache)

## Credits

This package is a modified version of [beego/cache](https://github.com/astaxie/beego/tree/master/cache).

## License

This project is under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for the full license text.

# Storer

[![GoDoc](https://godoc.org/github.com/ysmood/storer?status.svg)](http://godoc.org/github.com/ysmood/storer)
[![Build Status](https://travis-ci.org/ysmood/storer.svg?branch=master)](https://travis-ci.org/ysmood/storer)
[![codecov](https://codecov.io/gh/ysmood/storer/branch/master/graph/badge.svg)](https://codecov.io/gh/ysmood/storer)

On disk high performance object storage for golang. This project is based on my research of
the minimum interface to create an efficient indexable database with a key-value store.
It's a proof of concept for [pkg/kvstore](pkg/kvstore/interface.go).

## Features

- No table, no schema, no query language, just golang
- Manipulate records like normal list items in golang
- Query records with map and reduce in golang
- Complex indexing, such as compound indexes, object index, etc
- Complex transactions between different collections
- No database is perfect, use whatever backend that fits, by default it's badger
- Fluent api design, always get type suggestions from IDE

## Examples

## Development

The sugar files can be removed, it doesn't affect the core functions.

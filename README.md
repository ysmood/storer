# Storer

[![GoDoc](https://godoc.org/github.com/ysmood/storer?status.svg)](http://godoc.org/github.com/ysmood/storer)
[![Build Status](https://travis-ci.org/ysmood/storer.svg?branch=master)](https://travis-ci.org/ysmood/storer)
[![codecov](https://codecov.io/gh/ysmood/storer/branch/master/graph/badge.svg)](https://codecov.io/gh/ysmood/storer)

On-disk high-performance lightweight object storage for golang. This project is based on my research of
the minimum interface to create an efficient indexable database with a key-value store.
It's a proof of concept for [pkg/kvstore](pkg/kvstore/interface.go).

With this lib, it will be pretty easy to build SQL like DSL on top of it.

## Features

- No table, no schema, no query language, just golang
- Manipulate records like normal list items in golang
- Query records with `map` and `reduce` in golang
- Complex indexing, such as compound indexes, object index, etc
- Transactions between different collections
- No database is perfect, use whatever backend that fits, default is [badger](https://github.com/dgraph-io/badger)

## Examples

Check the [example file](examples_test.go).

## Benchmarks

This lib only added 3 layers above the underlying backend:

- minimum key prefixing, normally one byte per key, the algorithm is [here](github.com/ysmood/byframe)
- minimum data encoding, the encoding used for benchmarking is [msgpack](https://github.com/vmihailenco/msgpack)
- index items with an extra record, so need at least two gets to retrieve an item via its index

So theoretically, the performance should be nearly the same with bare get and set when data is small.

Run `go test -bench=.`, here's a sample output:

```txt
goos: darwin
goarch: amd64
pkg: github.com/ysmood/storer
BenchmarkBadgerSet-6           20000         72311 ns/op
BenchmarkSet-6                 20000         75317 ns/op
BenchmarkBadgerGet-6         1000000          1545 ns/op
BenchmarkGet-6               1000000          2389 ns/op
BenchmarkGetByIndex-6         300000          4196 ns/op
BenchmarkFilter-6              50000         23965 ns/op
PASS
ok      github.com/ysmood/storer    24.046s
```

As you can see, the real world benchmark reflects the theory,
GetByIndex is about 2 times slower than the direct Get.

## Development

Check the [travis file](.travis.yml) for the command for testing.

The sugar files can be removed, it doesn't affect the core functions.

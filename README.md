# Storer

[![GoDoc](https://godoc.org/github.com/ysmood/storer?status.svg)](http://godoc.org/github.com/ysmood/storer)
[![Build Status](https://travis-ci.org/ysmood/storer.svg?branch=master)](https://travis-ci.org/ysmood/storer)
[![codecov](https://codecov.io/gh/ysmood/storer/branch/master/graph/badge.svg)](https://codecov.io/gh/ysmood/storer)

On-disk high-performance lightweight object storage for golang. This project is based on my research of
the minimum interface to create an efficient indexable database with a key-value store.
It's a proof of concept for [pkg/kvstore](pkg/kvstore/interface.go).

It should be pretty easy to build low overhead random graph algorithm or SQL like DSL on top of it.

Normally you use badger as the backend at the early stage of your project, when you need to scale
you can seamlessly switch the backend to db like tikv or postgres.

## Features

- Manipulate records like normal list items in golang
- Complex indexing, such as compound indexes, object index, etc
- Transactions between collections and indexes
- No database is perfect, use whatever backend that fits, default is [badger](https://github.com/dgraph-io/badger)
- Graceful schema migration on the fly, no more stop the world migration, [how it works](pkg/typee/README.md)

## Examples

Check the [example file](examples_test.go). Also the [examples folder](examples).

## Benchmarks

This lib only added 3 layers above the underlying backend:

- minimum key prefixing, normally one byte per key, the algorithm is [here](https://github.com/ysmood/byframe)
- minimum data encoding, the encoding used for benchmarking is [msgpack](https://github.com/vmihailenco/msgpack)
- index items with an extra record, so need at least two gets to retrieve an item via its index

So theoretically, the performance should be nearly the same with bare get and set when data is small.

Run `go test -bench Benchmark -benchmem`, here's a sample output:

```txt
BenchmarkBadgerSet-6         	   20000	     72012 ns/op	    2539 B/op	      81 allocs/op
BenchmarkSet-6               	   20000	     73290 ns/op	    2756 B/op	      96 allocs/op
BenchmarkSetWithIndex-6      	   20000	     77295 ns/op	    3318 B/op	     120 allocs/op
BenchmarkBadgerGet-6         	 1000000	      1420 ns/op	     528 B/op	      13 allocs/op
BenchmarkGet-6               	 1000000	      1555 ns/op	     624 B/op	      19 allocs/op
BenchmarkBadgerPrefixGet-6   	  500000	      3861 ns/op	    1552 B/op	      34 allocs/op
BenchmarkGetByIndex-6        	  300000	      4187 ns/op	    1584 B/op	      43 allocs/op
BenchmarkFilter-6            	   50000	     23884 ns/op	    7870 B/op	     193 allocs/op
```

The ones named with badger are using badger direct to manipulate data, the ones after each badger benchmark
are the treatment group.

The benchmark shows badger's prefix-get has huge overhead over direct-get which doesn't make sense to me yet.

Overall the result is as expected, it shows the overhead of the key prefixing and data encoding decoding
can be ignored when compared with disk IO overhead.

## Development

Check the [travis file](.travis.yml) for the command for testing.

The sugar files can be removed, it doesn't affect the core functions.

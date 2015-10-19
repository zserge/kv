# kv

[![Build Status](https://travis-ci.org/zserge/kv.svg)](https://travis-ci.org/zserge/kv) [![](https://godoc.org/github.com/zserge/kv?status.svg)](http://godoc.org/github.com/zserge/kv)

An ultimately minimal persistent key-value store + LRU cache

## Benchmark

kv:

```
BenchmarkStoreGet-2                20000             59301 ns/op          84.32 MB/s
BenchmarkStoreSet-2                20000             68880 ns/op          72.59 MB/s
BenchmarkStoreCacheGet-2         3000000               404 ns/op       12347.96 MB/s
BenchmarkStoreCacheSet-2         2000000               715 ns/op        6983.40 MB/s
// Benchmarks below are copied from diskv tests
BenchmarkWrite__32B-2              50000             30470 ns/op           1.05 MB/s
BenchmarkWrite__1KB-2              50000             53329 ns/op          19.20 MB/s
BenchmarkWrite__4KB-2              30000             66490 ns/op          61.60 MB/s
BenchmarkWrite_10KB-2              20000            119928 ns/op          85.38 MB/s
BenchmarkRead__32B_NoCache-2        2000            523594 ns/op           0.06 MB/s
BenchmarkRead__1KB_NoCache-2        3000            641618 ns/op           1.60 MB/s
BenchmarkRead__4KB_NoCache-2        2000            733945 ns/op           5.58 MB/s
BenchmarkRead_10KB_NoCache-2        2000            803028 ns/op          12.75 MB/s
BenchmarkRead__32B_WithCache-2   1000000              1509 ns/op          21.20 MB/s
BenchmarkRead__1KB_WithCache-2   1000000              1221 ns/op         838.37 MB/s
BenchmarkRead__4KB_WithCache-2   1000000              1061 ns/op        3859.21 MB/s
BenchmarkRead_10KB_WithCache-2   1000000              1059 ns/op        9668.19 MB/s
```

Comparing to Diskv tested on the same machine:

```
BenchmarkWrite__32B_NoIndex-2      10000            113520 ns/op           0.28 MB/s
BenchmarkWrite__1KB_NoIndex-2      10000            111664 ns/op           9.17 MB/s
BenchmarkWrite__4KB_NoIndex-2      10000            118777 ns/op          34.48 MB/s
BenchmarkWrite_10KB_NoIndex-2      10000            136848 ns/op          74.83 MB/s
BenchmarkRead__32B_NoCache-2       50000             31648 ns/op           1.01 MB/s
BenchmarkRead__1KB_NoCache-2       50000             34720 ns/op          29.49 MB/s
BenchmarkRead__4KB_NoCache-2       20000             70228 ns/op          58.32 MB/s
BenchmarkRead_10KB_NoCache-2       10000            103389 ns/op          99.04 MB/s
BenchmarkRead__32B_WithCache-2    200000              7671 ns/op           4.17 MB/s
BenchmarkRead__1KB_WithCache-2    200000              9299 ns/op         110.11 MB/s
BenchmarkRead__4KB_WithCache-2     30000             33615 ns/op         121.85 MB/s
BenchmarkRead_10KB_WithCache-2     20000             93018 ns/op         110.09 MB/s
```

I don't say that `kv` is faster or better, but it's fast enough for a simple
key-value file-based storage.
